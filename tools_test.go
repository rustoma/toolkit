package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)

	if len(s) != 10 {
		t.Error("wrong length random string returned")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{
		name:          "allowed no rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    false,
		errorExpected: false,
	},
	{
		name:          "allowed rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    true,
		errorExpected: false,
	},
	{
		name:          "not allowed",
		allowedTypes:  []string{"image/jpeg"},
		renameFile:    false,
		errorExpected: true,
	},
}

var slugifyTests = []struct {
	name           string
	entryString    string
	expectedOutput string
	errorExpected  bool
}{
	{
		name:           "valid string",
		entryString:    "Hello World",
		expectedOutput: "hello-world",
		errorExpected:  false,
	},
	{
		name:           "valid string",
		entryString:    "Example string !! To test ??",
		expectedOutput: "example-string-to-test",
		errorExpected:  false,
	},
	{
		name:           "valid string",
		entryString:    "aąbcć dęefł",
		expectedOutput: "a-bc-d-ef",
		errorExpected:  false,
	},
	{
		name:           "not valid string",
		entryString:    "",
		expectedOutput: "",
		errorExpected:  true,
	},
	{
		name:           "japanese not valid string",
		entryString:    "こんにちは世界",
		expectedOutput: "",
		errorExpected:  true,
	},
	{
		name:           "japanese valid string",
		entryString:    "helloこんにちは世界world",
		expectedOutput: "hello-world",
		errorExpected:  false,
	},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			// create the form data field 'file'

			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		//read from the pipe which receives data

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: Expected file to exist: %s", e.name, err.Error())
			}

			//clean up
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but not received", e.name)
		}

		wg.Wait()

	}
}

func TestTools_UploadOneFile(t *testing.T) {

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()

		// create the form data field 'file'

		part, err := writer.CreateFormFile("file", "./testdata/img.png")
		if err != nil {
			t.Error(err)
		}

		f, err := os.Open("./testdata/img.png")
		if err != nil {
			t.Error(err)
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			t.Error("error decoding image", err)
		}

		err = png.Encode(part, img)
		if err != nil {
			t.Error(err)
		}
	}()

	//read from the pipe which receives data

	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools

	uploadedFiles, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", err.Error())
	}

	//clean up
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName))
}

func TestTools_CreateDirIfNotExist(t *testing.T) {

	var testTool Tools

	err := testTool.CreateDirIfNotExist("./testdata/create-dir-if-not-exist")
	if err != nil {
		t.Error(err)
	}

	err = testTool.CreateDirIfNotExist("./testdata/create-dir-if-not-exist")
	if err != nil {
		t.Error(err)
	}

	//clean up
	_ = os.Remove("./testdata/create-dir-if-not-exist")

}

func TestTools_Slugify(t *testing.T) {

	var toolkit Tools

	for _, e := range slugifyTests {
		slug, err := toolkit.Slugify(e.entryString)

		if err != nil && !e.errorExpected {
			t.Errorf("%s error received when none expected: %s", e.name, err.Error())
		}

		if !e.errorExpected && slug != e.expectedOutput {
			t.Errorf("%s: Slugify return %s, but expected %s", e.name, slug, e.expectedOutput)
		}
	}

}
