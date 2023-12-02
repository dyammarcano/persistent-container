package main

import "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

func myblob() {
	clienr, err := azblob.NewBlobClient("https://mystorageaccount.blob.core.windows.net", azblob.NewAnonymousCredential(), nil)
}
