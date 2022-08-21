#!/bin/bash
gcloud functions deploy updateplayerreport --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=UpdatePlayerReport --trigger-topic=playerreport
