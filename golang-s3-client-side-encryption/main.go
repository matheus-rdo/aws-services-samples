package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/matheushr97/aws-services-samples/settings"
)

func main() {
	start := time.Now()

	err := uploadEncrypted()
	//err := downloadEncrypted()
	if err != nil {
		fmt.Println("ERRO: " + err.Error())
	}

	fmt.Println("Duração: " + time.Since(start).String())
}

// ENCRYPTED

func uploadEncrypted() error {
	sess := session.New(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	// This is our key wrap handler, used to generate cipher keys and IVs for
	// our cipher builder. Using an IV allows more “spontaneous” encryption.
	// The IV makes it more difficult for hackers to use dictionary attacks.
	// The key wrap handler behaves as the master key. Without it, you can’t
	// encrypt or decrypt the data.
	var matDesc s3crypto.MaterialDescription
	cipherDataGenerator := s3crypto.NewKMSContextKeyGenerator(kms.New(sess), settings.KmsKeyArn, matDesc)

	// This is our content cipher builder, used to instantiate new ciphers
	// that enable us to encrypt or decrypt the payload.
	contentCipherBuilder := s3crypto.AESGCMContentCipherBuilderV2(cipherDataGenerator)
	// Let's create our crypto client!
	encryptionClient, err := s3crypto.NewEncryptionClientV2(sess, contentCipherBuilder)
	if err != nil {
		return err
	}

	file, err := os.Open(settings.FileName)
	if err != nil {
		return err
	}

	objectKey := file.Name() + ".encrypted"
	input := &s3.PutObjectInput{
		Bucket: &settings.BucketName,
		Key:    &objectKey,
		Body:   file,
	}

	_, err = encryptionClient.PutObject(input)
	// What to expect as errors? You can expect any sort of S3 errors, http://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html.
	// The s3crypto client can also return some errors:
	//  * MissingCMKIDError - when using AWS KMS, the user must specify their key's ARN
	if err == nil {
		fmt.Println("Object sucessfully uploaded")
	}

	return err
}

func downloadEncrypted() error {
	sess := session.New(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	registry := s3crypto.NewCryptoRegistry()
	// Register required content decryption algorithms
	if err := s3crypto.RegisterAESGCMContentCipher(registry); err != nil {
		return err
	}

	// Register required key wrapping algorithms
	// Use RegisterKMSContextWrapWithCMK to limit the KMS Decrypt to a single CMK
	if err := s3crypto.RegisterKMSContextWrapWithCMK(registry, kms.New(sess), settings.KmsKeyArn); err != nil {
		panic(err)
	}

	decryptionClient, err := s3crypto.NewDecryptionClientV2(sess, registry)
	if err != nil {
		return err
	}

	objectKey := settings.FileName + ".encrypted"
	input := &s3.GetObjectInput{
		Bucket: &settings.BucketName,
		Key:    &objectKey,
	}

	result, err := decryptionClient.GetObject(input)
	// Aside from the S3 errors, here is a list of decryption client errors:
	//   * InvalidWrapAlgorithmError - returned on an unsupported Wrap algorithm
	//   * InvalidCEKAlgorithmError - returned on an unsupported CEK algorithm
	//   * V1NotSupportedError - the SDK doesn’t support v1 because security is an issue for AES ECB
	// These errors don’t necessarily mean there’s something wrong. They just tell us we couldn't decrypt some data.
	// Users can choose to log this and then continue decrypting the data that they can, or simply return the error.
	if err != nil {
		return err
	}

	outputFile := settings.FileName + ".decrypted"
	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	n, err := io.Copy(output, result.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Downloaded %d bytes \n", n)
	return nil
}

// UNENCRYPTED

func upload() error {
	sess := session.New(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	file, err := os.Open(settings.FileName)
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(sess)
	objectKey := file.Name() + ".unencrypted"
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: &settings.BucketName,
		Key:    &objectKey,
		Body:   file,
	})
	if err != nil {
		return err
	}

	fmt.Println("Object sucessfully uploaded")
	return nil
}

func download() error {
	sess := session.New(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	outputFile := settings.FileName + ".unencrypted"
	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	downloader := s3manager.NewDownloader(sess)
	n, err := downloader.Download(output, &s3.GetObjectInput{
		Bucket: &settings.BucketName,
		Key:    &outputFile,
	})

	fmt.Printf("Downloaded %d bytes \n", n)
	return nil
}
