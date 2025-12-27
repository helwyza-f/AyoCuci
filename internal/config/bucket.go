package config

import(
	"os"
)

var(
	// Credential AWS S3
	AWS_ACCESS_KEY_ID		string	
	AWS_SECRET_ACCESS_KEY	string	 
	AWS_REGION				string	
	AWS_ENDPOINT			string	 
	AWS_BUCKET_NAME			string	 
)

func InitBucketVars() {
	AWS_ACCESS_KEY_ID		= os.Getenv("AWS_ACCESS_KEY_ID"+Prefix)
	AWS_SECRET_ACCESS_KEY	= os.Getenv("AWS_SECRET_ACCESS_KEY"+Prefix)
	AWS_REGION				= os.Getenv("AWS_REGION"+Prefix)
	AWS_ENDPOINT			= os.Getenv("AWS_ENDPOINT"+Prefix)
	AWS_BUCKET_NAME			= os.Getenv("AWS_BUCKET_NAME"+Prefix)
}