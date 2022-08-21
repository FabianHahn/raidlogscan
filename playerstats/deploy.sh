#!/bin/bash
gcloud functions deploy playerstats --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=PlayerStats --trigger-http --allow-unauthenticated
