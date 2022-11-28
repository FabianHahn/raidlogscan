#!/bin/bash
gcloud storage cp source.zip gs://raidlogscan_sources/source.zip

gcloud functions deploy accountstats --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=AccountStats --trigger-http --allow-unauthenticated
gcloud functions deploy claimaccount --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=ClaimAccount --trigger-http --allow-unauthenticated
gcloud functions deploy playerstats --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=PlayerStats --trigger-http --allow-unauthenticated
gcloud functions deploy oauth2login --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=Oauth2Login --trigger-http --allow-unauthenticated
gcloud functions deploy oauth2callback --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=Oauth2Callback --trigger-http --allow-unauthenticated
gcloud functions deploy scanuserreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=ScanUserReports --trigger-http --allow-unauthenticated
gcloud functions deploy scanrecentcharacterreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=ScanRecentCharacterReports --trigger-http --allow-unauthenticated

gcloud functions deploy coraideraccountclaim --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=CoraiderAccountClaim --trigger-topic=coraideraccountclaim
gcloud functions deploy fetchguildreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=FetchGuildReports --trigger-topic=guildreports
gcloud functions deploy fetchreport2 --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=FetchReport --trigger-topic=report
gcloud functions deploy updateplayerreport --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=UpdatePlayerReport --trigger-topic=playerreport
gcloud functions deploy fetchuserreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=FetchUserReports --trigger-topic=userreports
gcloud functions deploy fetchrecentcharacterreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=FetchRecentCharacterReports --trigger-topic=recentcharacterreports
