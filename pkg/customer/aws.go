/*
Copyright © 2024 Mattias Gees mattias.gees@venafi.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package customer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Retrieves a file from S3 and shows that file to the customer
func (c *CustomerService) awsRetrievalHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the AWS Retrieval Handler from %s", r.RemoteAddr)

	ctx := context.Background()

	// Load AWS configuration.
	// Region is determined by AWS SDK from environment (AWS_REGION or AWS_DEFAULT_REGION).
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load AWS config: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a new S3 client.
	client := s3.NewFromConfig(cfg)

	// Retrieve a file from S3
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.awsBucket),
		Key:    aws.String(c.awsFilePath),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get object: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the content of the retrieved file from S3
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read object content", http.StatusInternalServerError)
		return
	}

	// Showcase the content of the retrieved file to the customer
	tmpl := template.Must(template.New("display").Parse(`
		<html>
		<head><title>S3 File Content</title></head>
		<body>
				<h1>S3 File Content</h1>
				<pre>{{.}}</pre>
		</body>
		</html>
`))

	err = tmpl.Execute(w, string(content))
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// Writes a file to S3 and shows the success to the customer
func (c *CustomerService) awsPutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the AWS Put Handler from %s", r.RemoteAddr)

	ctx := context.Background()

	// Load AWS configuration.
	// Region is determined by AWS SDK from environment (AWS_REGION or AWS_DEFAULT_REGION).
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load AWS config: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a new S3 client.
	client := s3.NewFromConfig(cfg)
	reader := bytes.NewReader([]byte("This is a test to write to an S3 bucket"))

	// Write a file to S3
	result, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.awsBucket),
		Key:    aws.String(c.awsFilePath),
		Body:   reader,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to upload %q to %q, %v", c.awsFilePath, c.awsBucket, err), http.StatusInternalServerError)
		return
	}

	// Tell the customer we have uploaded a file to S3 and add some information to where on S3 we have uploaded it.
	fmt.Fprintf(w, "Successfully uploaded %q to %q\n", c.awsFilePath, c.awsBucket)
	fmt.Fprintf(w, "The uploaded content is: %v", result)
}
