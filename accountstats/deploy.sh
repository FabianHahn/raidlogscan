#!/bin/bash
gcloud functions deploy accountstats --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=AccountStats --trigger-http --allow-unauthenticated
