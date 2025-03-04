package function

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/mholt/archives"
)

func FormFileToBytes(ctx echo.Context, formFile *multipart.FileHeader) ([]byte, error) {
	src, err := formFile.Open()
	if err != nil {
		return nil, ctx.String(http.StatusInternalServerError, "Unable to open the file: "+err.Error())
	}
	defer src.Close()

	// Read the entire file into memory (do not save on disk)
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, ctx.String(http.StatusInternalServerError, "Error reading the file: "+err.Error())
	}

	return fileBytes, nil

}

func ProcessRequestData(ctx echo.Context) (fileBytes []byte, runtime string, name string, envVars []EnvVar, error error) {
	// Retrieve the uploaded file from the "file" field.
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, "", "", nil, ctx.String(http.StatusBadRequest, "Error retrieving the file: "+err.Error())
	}

	fileBytes, err = FormFileToBytes(ctx, fileHeader)
	if err != nil {
		return nil, "", "", nil, err
	}
	// Retrieve other form fields.
	runtimeField := ctx.FormValue("runtime")
	nameField := ctx.FormValue("name")
	envVarsStr := ctx.FormValue("env_vars")

	// Parse the JSON array of environment variables.
	if envVarsStr != "" {
		if err := json.Unmarshal([]byte(envVarsStr), &envVars); err != nil {
			return nil, "", "", nil, ctx.String(http.StatusBadRequest, "Error parsing env_vars: "+err.Error())
		}
	}

	return fileBytes, runtimeField, nameField, envVars, nil
}

func UnknownToTar(fileBytes []byte) ([]byte, error) {
	// identify format
	format, stream, err := archives.Identify(context.TODO(), "user-code", bytes.NewReader(fileBytes))
	if err != nil {
		log.Fatal(err)
	}

	switch format.MediaType() {
	case "application/x-7z-compressed":
		return nil, fmt.Errorf("7z format is not supported")
	case "application/zip":
		tar, err := ZipToTar(stream)
		if err != nil {
			return nil, fmt.Errorf("error converting zip to tar: %w", err)
		}
		return tar, nil

	default:
		return nil, fmt.Errorf("unsupported format: %v", format.MediaType())
	}
}

func ZipToTar(stream io.Reader) ([]byte, error) {
	// Read the entire ZIP archive into memory.
	zipData, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading zip stream: %w", err)
	}

	// Create a new zip.Reader using a bytes.Reader (which implements io.ReaderAt).
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("creating zip reader: %w", err)
	}

	// Create a buffer to hold the tar archive.
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)
	defer tarWriter.Close()

	// Iterate over each file in the ZIP archive.
	for _, zipFile := range zipReader.File {
		// Open the file within the ZIP archive.
		f, err := zipFile.Open()
		if err != nil {
			return nil, fmt.Errorf("opening zip file %s: %w", zipFile.Name, err)
		}

		// Create a tar header from the file info.
		// The second parameter (link string) is only used for symlinks.
		header, err := tar.FileInfoHeader(zipFile.FileInfo(), zipFile.Name)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("creating tar header for %s: %w", zipFile.Name, err)
		}
		// Ensure the header name is set to the original file name.
		header.Name = zipFile.Name

		// Write the header into the tar archive.
		if err := tarWriter.WriteHeader(header); err != nil {
			f.Close()
			return nil, fmt.Errorf("writing tar header for %s: %w", zipFile.Name, err)
		}

		// Copy the file data from the ZIP entry to the TAR archive.
		if _, err := io.Copy(tarWriter, f); err != nil {
			f.Close()
			return nil, fmt.Errorf("copying file data for %s: %w", zipFile.Name, err)
		}
		f.Close()
	}

	// Ensure the tar writer is closed (deferred above) so all data is flushed.
	return buf.Bytes(), nil
}

// func isTar(tarData []byte) (bool, error) {
// 	// identify format
// 	format, _, err := archives.Identify(context.TODO(), "user-code", bytes.NewReader(tarData))
// 	if err != nil {
// 		log.Fatal(err)
// 		return false, err
// 	}

// 	if format.MediaType() == "application/x-tar" {
// 		return true, nil
// 	}
// 	return false, nil
// }

// InjectDockerfile injects the local Dockerfile (from disk) into the provided tar archive.
// It returns a new tar archive as a []byte that contains all original entries plus the Dockerfile.
func InjectDockerfile(tarData []byte) ([]byte, error) {
	// Create a buffer for the new tar archive.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Create a tar reader for the existing tar archive.
	tr := tar.NewReader(bytes.NewReader(tarData))

	// Copy all entries from the original tar archive.
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive.
		}
		if err != nil {
			tw.Close()
			return nil, fmt.Errorf("error reading original tar: %w", err)
		}

		// Write the header to the new tar archive.
		if err := tw.WriteHeader(header); err != nil {
			tw.Close()
			return nil, fmt.Errorf("error writing header for %s: %w", header.Name, err)
		}

		// For regular files, copy the file content.
		if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
			if _, err := io.Copy(tw, tr); err != nil {
				tw.Close()
				return nil, fmt.Errorf("error copying data for %s: %w", header.Name, err)
			}
		}
	}

	// Inject the Dockerfile into the archive.
	dockerFilePath := "node/22.14.0/Dockerfile"

	// Read Dockerfile's file info and content.
	fileInfo, err := os.Stat(dockerFilePath)
	if err != nil {
		tw.Close()
		return nil, fmt.Errorf("error stating Dockerfile: %w", err)
	}
	dockerFile, err := os.Open(dockerFilePath)
	if err != nil {
		tw.Close()
		return nil, fmt.Errorf("error opening Dockerfile: %w", err)
	}
	defer dockerFile.Close()

	dockerData, err := io.ReadAll(dockerFile)
	if err != nil {
		tw.Close()
		return nil, fmt.Errorf("error reading Dockerfile: %w", err)
	}

	// Create a tar header for the Dockerfile.
	header, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		tw.Close()
		return nil, fmt.Errorf("error creating header for Dockerfile: %w", err)
	}
	// Set the header name to "Dockerfile" so it appears at the root.
	header.Name = "Dockerfile"
	// Optionally update the modification time (or keep fileInfo.ModTime())
	header.ModTime = time.Now()

	// Write the Dockerfile header and content.
	if err := tw.WriteHeader(header); err != nil {
		tw.Close()
		return nil, fmt.Errorf("error writing Dockerfile header: %w", err)
	}
	if _, err := tw.Write(dockerData); err != nil {
		tw.Close()
		return nil, fmt.Errorf("error writing Dockerfile data: %w", err)
	}

	// Close the tar writer to flush all data.
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("error closing tar writer: %w", err)
	}

	return buf.Bytes(), nil
}
