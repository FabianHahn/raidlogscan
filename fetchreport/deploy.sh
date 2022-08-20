#!/bin/bash
gcloud functions deploy fetchreport --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=FetchReport --trigger-http --allow-unauthenticated
