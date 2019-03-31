package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/etherlabsio/hls/pkg/hls"
)

func Test_lambdaInvoke(t *testing.T) {
	tests := []struct {
		name                    string
		Bucket                  string
		KeyPattern              string
		start                   int
		stop                    int
		DRMKey                  []byte
		DRMInitializationVector string
		wantErr                 bool
	}{
		{
			"Basic functionality",
			os.Getenv("AWS_BUCKET"),
			"recordings/hls-test/out%04d.ts",
			12,
			19,
			[]byte{122, 108, 181, 101, 231, 152, 205, 196, 127, 105, 169, 155, 109, 124, 120, 213},
			"51532a902859d8b493b6db31a1d1cfc5",
			false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			// Create Lambda service client
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := lambda.New(sess, &aws.Config{})

			for k := tt.start; k <= tt.stop; k++ {
				key := fmt.Sprintf(tt.KeyPattern, k)
				request := hls.TranscodeEvent{tt.Bucket, key,
					tt.DRMKey, tt.DRMInitializationVector, []string{"720p", "360p", "480p"}}

				payload, err := json.Marshal(request)

				if (err != nil) != tt.wantErr {
					fmt.Println("Error marshalling testBasicLambdaFunctionality request")
					return
				}

				_, err = client.Invoke(&lambda.InvokeInput{FunctionName: aws.String("ether-hls-multirate-transcoder"), Payload: payload, InvocationType: aws.String("Event")})

				if (err != nil) != tt.wantErr {
					fmt.Println("Error calling MyGetItemsFuntestBasicLambdaFunctionalityction")
					return
				}
			}
		})
	}
}
