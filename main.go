package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

func main() {

	// the folder to examine should be passed as a command line argument or parameter.  If no parameters supplied, then default to current folder
	folderPath := getAndCheckFolderPath()
	// get a list of the file extensions we care about (jpg, png, etc...)
	imagefileExtensions := createMapOfRecognisedImageFileExtensions()

	// initialise flags and counters
	zeroTime := time.Time{}
	imageFileCount := 0
	actionedImages := 0

	// now scan the files in the folderpath
	// and and look for valid imagefile extensions
	fmt.Println("Scanning folder ", folderPath, " for JPG, JPEG, TIFF, and PNG files...")
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		// an unexpected error occurred while obtaining
		// the list of files in the folder path...
		log.Fatal(err)
	}

	for _, f := range files {

		fileName := f.Name()
		//get the file extension and convert it to lowercase
		fileExtension := strings.ToLower(filepath.Ext(fileName))

		//is this a file extension we care about?
		if _, isImage := imagefileExtensions[fileExtension]; isImage {

			//yes, we recognise the file extension.  Does this file have an EXIF creation date-time tag?
			timestamp := getFileExifCreateDate(folderPath + "/" + fileName)
			imageFileCount++
			if timestamp == zeroTime {
				fmt.Println(fileName, ": No EXIF creation date/time")
			} else {
				// we have an image with EXIF timestamp data so take action!
				processTimeStampedImage(folderPath, fileName, timestamp)
				actionedImages++
			}
		}
	}
	fmt.Println("Scan complete!", imageFileCount, " Image files found, ", actionedImages, " had a valid timestamp and were moved into subfolders.")
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func processTimeStampedImage(filepath string, fileName string, timestamp time.Time) {

	//we have a timestamp to work with
	// first, check if a folder for this timestamp has already been created?
	subfoldername := timestamp.Format("2006_01_02")
	err := CreateDirIfNotExist(filepath + "/" + subfoldername)
	if err != nil {
		fmt.Println("Error! Unable to create sub-folder! Giving Up! The Error Message is: ", err)
		os.Exit(2)
	}

	oldpath := filepath + "/" + fileName
	newpath := filepath + "/" + subfoldername + "/" + fileName

	err = os.Rename(oldpath, newpath)
	if err != nil {
		fmt.Println("Error! ", err)
		os.Exit(3)
	}
	fmt.Println(fileName, "moved to", subfoldername)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func getFileExifCreateDate(filespec string) time.Time {

	f, err := os.Open(filespec)
	defer f.Close()

	if err != nil {
		fmt.Println("Error opening file: ", filespec, ". Error: ", err)
		return time.Time{}
	}

	exifData, err := exif.Decode(f)
	if err != nil {
		return time.Time{}
	}

	createdate, _ := exifData.DateTime()
	return createdate
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func getAndCheckFolderPath() string {
	argsWithoutProg := os.Args[1:]

	// default to current folder unless supplied as a command line argument
	folderpath := "./"
	if len(argsWithoutProg) != 0 {
		folderpath = argsWithoutProg[0]
	}

	//check if the folder exists
	src, err := os.Stat(folderpath)
	if err != nil {
		fmt.Println("Folder: ", folderpath, " not found")
		os.Exit(1)
	}

	//check if the folderpath is indeed a folder and not a file
	if !src.IsDir() {
		fmt.Println(folderpath, " is not a folder")
		os.Exit(1)
	}
	return folderpath
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func createMapOfRecognisedImageFileExtensions() map[string]string {
	//we have been given a valid folder, now scan it
	// for files with extensions:
	//
	//   jpg, jpeg
	//   png
	//   tiff
	//
	// create a map of file extensions that can contain EXIF data
	//
	extns := map[string]string{}
	extns[".jpeg"] = "Y"
	extns[".jpg"] = "Y"
	extns[".tiff"] = "Y"
	extns[".png"] = "Y"

	return extns
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
