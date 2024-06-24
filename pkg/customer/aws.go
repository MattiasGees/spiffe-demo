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
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Retrieves a file from S3 and shows that file to the customer
func (c *CustomerService) awsRetrievalHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the AWS Retrieval Handler from %s", r.RemoteAddr)

	// Setup a session to AWS.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(c.awsRegion),
	})
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Create a new S3 instance.
	svc := s3.New(sess)
	// Retrieve a file from S3
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.s3Bucket),
		Key:    aws.String(c.s3Filepath),
	})
	if err != nil {
		http.Error(w, "Failed to get object", http.StatusInternalServerError)
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
	verbose := true

	// Setup a session to AWS.
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(c.awsRegion),
		CredentialsChainVerboseErrors: &verbose,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create session, %v", err), http.StatusInternalServerError)
		return
	}

	// Create a new S3 instance.
	svc := s3.New(sess)
	reader := bytes.NewReader([]byte("This is a test to write to an S3 bucket"))

	// Write a file to S3
	result, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(c.s3Bucket),
		Key:    aws.String(c.s3Filepath),
		Body:   reader,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to upload %q to %q, %v", c.s3Filepath, c.s3Bucket, err), http.StatusInternalServerError)
		return
	}

	// Tell the customer we have uploaded a file to S3 and add some information to where on S3 we have uploaded it.
	fmt.Fprintf(w, "Successfully uploaded %q to %q\n", c.s3Filepath, c.s3Bucket)
	fmt.Fprintf(w, "The uploaded content is: %s", result)
}
