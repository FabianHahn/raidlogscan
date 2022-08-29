#!/bin/bash
gcloud functions deploy claimaccount --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=ClaimAccount --trigger-http --allow-unauthenticated
