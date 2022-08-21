#!/bin/bash
gcloud functions deploy fetchguildreports --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=FetchGuildReports --trigger-topic=guildreports
