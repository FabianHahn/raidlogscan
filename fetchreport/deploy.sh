#!/bin/bash
gcloud functions deploy fetchreport2 --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=FetchReport --trigger-topic=report
