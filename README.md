#IAP_PROXY

This is a proxy server designed to be run locally for accessing resources behind Google's Identity Aware proxy inspired by [iap_curl](https://github.com/b4b4r07/iap_curl/)


You must set three environment variables.

1. `export GOOGLE_APPLICATION_CREDENTIALS="path to your service account json"`
2. `export IAP_CLIENT_ID="client id"` see https://cloud.google.com/iap/docs/authentication-howto
3. `export IAP_HOST ="URL_IAP_SERVICE"`