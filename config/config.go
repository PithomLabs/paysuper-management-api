package config

import (
	"github.com/kelseyhightower/envconfig"
)

const (
	DefaultJwtSignAlgorithm = "RS256"
)

type Auth1 struct {
	Issuer       string `envconfig:"AUTH1_ISSUER" required:"true" default:"https://dev-auth1.tst.protocol.one"`
	ClientId     string `envconfig:"AUTH1_CLIENTID" required:"true"`
	ClientSecret string `envconfig:"AUTH1_CLIENTSECRET" required:"true"`
	RedirectUrl  string `envconfig:"AUTH1_REDIRECTURL" required:"true"`
}

type S3 struct {
	AccessKeyId string `envconfig:"S3_ACCESS_KEY" required:"true"`
	SecretKey   string `envconfig:"S3_SECRET_KEY" required:"true"`
	Endpoint    string `envconfig:"S3_ENDPOINT" required:"true"`
	BucketName  string `envconfig:"S3_BUCKET_NAME" required:"true"`
	Region      string `envconfig:"S3_REGION" default:"us-west-2"`
	Secure      bool   `envconfig:"S3_SECURE" default:"false"`
}

type S3Report struct {
	AccessKeyId string `envconfig:"S3REPORT_ACCESS_KEY" required:"true"`
	SecretKey   string `envconfig:"S3REPORT_SECRET_KEY" required:"true"`
	Endpoint    string `envconfig:"S3REPORT_ENDPOINT" required:"true"`
	BucketName  string `envconfig:"S3REPORT_BUCKET_NAME" required:"true"`
	Region      string `envconfig:"S3REPORT_REGION" default:"us-west-1"`
	Secure      bool   `envconfig:"S3REPORT_SECURE" default:"false"`
}

type Config struct {
	Auth1
	S3
	S3Report

	HttpScheme  string `envconfig:"HTTP_SCHEME" default:"https"`
	Environment string `envconfig:"ENVIRONMENT" default:"test"`
}

func NewConfig() (error, *Config) {
	var err error

	config := Config{}

	if err = envconfig.Process("", &config); err != nil {
		return err, nil
	}

	return nil, &config
}
