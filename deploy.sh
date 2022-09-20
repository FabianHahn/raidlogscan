#!/bin/bash
gcloud functions deploy accountstats --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=AccountStats --trigger-http --allow-unauthenticated
gcloud functions deploy claimaccount --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=ClaimAccount --trigger-topic=claimaccount
gcloud functions deploy claimaccount --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=ClaimAccount --trigger-http --allow-unauthenticated
gcloud functions deploy fetchguildreports --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=FetchGuildReports --trigger-topic=guildreports
gcloud functions deploy fetchreport2 --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=FetchReport --trigger-topic=report
gcloud functions deploy playerstats --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=PlayerStats --trigger-http --allow-unauthenticated
gcloud functions deploy updateplayerreport --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=UpdatePlayerReport --trigger-topic=playerreport
