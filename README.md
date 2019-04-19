
# IAP_PROXY
This is a proxy server designed to be run locally for accessing resources behind Google's Identity Aware proxy inspired by [iap_curl](https://github.com/b4b4r07/iap_curl/)


You must set three environment variables.

2. `export IAP_CLIENT_ID="client id"` see https://cloud.google.com/iap/docs/authentication-howto
3. `export IAP_HOST ="URL_IAP_SERVICE"`


If you are using the program for the first time, start the program `iap_proxy --cred <path to google service account JWT>` this will save the service account in local keyring and prompt you to delete it.