AWS_ACCESS_KEY_ID=$(shell aws configure get aws_access_key_id --profile ${AWS_PROFILE})
AWS_SECRET_ACCESS_KEY=$(shell aws configure get aws_secret_access_key --profile ${AWS_PROFILE})
AWS_REGION=$(shell aws configure get region --profile ${AWS_PROFILE})

SLACK_WEBHOOK_URL="https://hooks.slack.com/services/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"
BRANCH=$(shell git rev-parse HEAD || echo -e '$CI_COMMIT_SHA')

pre-deploy-notify:
	@curl -X POST --data-urlencode 'payload={"text": "[${ENVIRONMENT}] [${BRANCH}] ${USER}: ${ARTIFACT} is being deployed"}' \
                 ${SLACK_WEBHOOK_URL}

post-deploy-notify:
	@curl -X POST --data-urlencode 'payload={"text": "[${ENVIRONMENT}] [${BRANCH}] ${USER}: ${ARTIFACT} is deployed"}' \
                 ${SLACK_WEBHOOK_URL}


build_lambda:
	GOOS=linux GOARCH=amd64 go build -tags debug -v -o ./bin/${LAMBDA_ZIP} ./cmd/${ARTIFACT} 
	zip ./bin/${LAMBDA_ZIP}-lambda.zip ./bin/${LAMBDA_ZIP}

deploy_lambda: build_lambda
	$(MAKE) pre-deploy-notify
	    aws s3 cp --profile production bin/${LAMBDA_ZIP}-lambda.zip s3://io.etherlabs.artifacts/${ENVIRONMENT}/${LAMBDA_ZIP}-lambda.zip
		aws lambda update-function-code --region ${AWS_REGION} --function-name ${LAMBDA_FUNCTION} --s3-bucket io.etherlabs.artifacts --s3-key ${ENVIRONMENT}/${LAMBDA_ZIP}-lambda.zip
	$(MAKE) post-deploy-notify

deploy_hls_multirate_transcoder_staging:
	$(MAKE) deploy_lambda ENVIRONMENT=staging ARTIFACT=hls-multirate-transcoder LAMBDA_ZIP=hls-multirate-transcoder LAMBDA_FUNCTION=ether-hls-multirate-transcoder \
		REGION=$(AWS_REGION) AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY)

deploy_hls_multirate_transcoder_production:
	$(MAKE) deploy_lambda ENVIRONMENT=production ARTIFACT=hls-multirate-transcoder LAMBDA_ZIP=hls-multirate-transcoder LAMBDA_FUNCTION=ether-hls-multirate-transcoder \
		REGION=$(AWS_REGION) AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY)

deploy_hls_multirate_transcoder_staging2:
	$(MAKE) deploy_lambda ENVIRONMENT=staging2 ARTIFACT=hls-multirate-transcoder LAMBDA_ZIP=hls-multirate-transcoder LAMBDA_FUNCTION=ether-hls-multirate-transcoder \
		REGION=$(AWS_REGION) AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) AWS_SECRET_ACCESS_KEY=$(_AWS_SECRET_ACCESS_KEY)
