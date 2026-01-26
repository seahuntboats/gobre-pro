package libreoffice

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	storagePath = "/dev/shm/"
)

func HandleConvertFile(
	ogFileType string,
	newFileType string,
	fileData []byte,
) ([]byte, error) {
	if len(ogFileType) == 0 || len(newFileType) == 0 {
		return nil, status.Error(codes.NotFound, "Invlaid file types")
	}

	randBytes := strconv.Itoa(rand.Int())

	filePtr, _ := os.OpenFile(
		storagePath+randBytes,
		os.O_WRONLY|os.O_CREATE,
		0644,
	)

	bodyBytes := fileData

	_, fileWriteError := filePtr.Write(bodyBytes)
	if fileWriteError != nil {
		fmt.Println("Error writing file: ", fileWriteError)
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprint(fileWriteError))
	}
	fileCloseErr := filePtr.Close()

	if fileCloseErr != nil {
		fmt.Println("Error closing file: ", fileCloseErr)
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprint(fileCloseErr))
	}

	//Attempt to convert raw file to the original format to ensure its a valid file
	//This also normally fixes a few formatting issues from dynamic file generators.
	libreofficeCmdError := runLibreoffice(ogFileType, newFileType, randBytes)
	if libreofficeCmdError != nil {
		fmt.Println("Error running libreoffice command: ", libreofficeCmdError)
		return nil, status.Error(codes.Unknown, fmt.Sprint(libreofficeCmdError))
	}

	//Convert to new file
	data, readFileError := os.ReadFile(storagePath + randBytes + "." + newFileType)
	if readFileError != nil {
		fmt.Println("Error reading converted file: ", readFileError)
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprint(readFileError))
	}

	//Clean out files and libreoffice profile
	cleanDevShm(randBytes, newFileType)
	return data, nil
}

func runLibreoffice(
	ogFileType string,
	newFileType string,
	randBytes string,
) error {
	var libreofficeOptions string
	if ogFileType == "pdf" {
		libreofficeOptions = "--infilter='writer_pdf_import'"
	}
	libreofficeCmd := exec.Command(
		"bash",
		"-c",
		"libreoffice "+libreofficeOptions+
			" -env:UserInstallation=file://"+storagePath+randBytes+"_lo "+
			"--headless --convert-to "+newFileType+" "+randBytes,
	)

	libreofficeCmd.Dir = "/dev/shm"
	return libreofficeCmd.Run()
}

func cleanDevShm(fileName string, newFileType string) {
	ogRemoveError := os.Remove(storagePath + fileName)

	if ogRemoveError != nil {
		fmt.Println("Error removing original file: ", ogRemoveError)
	}

	newRemoveError := os.Remove(storagePath + fileName + "." + newFileType)

	if newRemoveError != nil {
		fmt.Println("Error removing converted file: ", newRemoveError)
	}

	configRemoveError := os.RemoveAll(storagePath + fileName + "_lo")

	if configRemoveError != nil {
		fmt.Println("Error removing libreoffice config: ", configRemoveError)
	}
}
