#!/bin/bash
gcloud storage cp source.zip gs://raidlogscan_sources/source.zip

gcloud functions deploy accountstats --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=AccountStats --trigger-http --allow-unauthenticated
gcloud functions deploy claimaccount --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=ClaimAccount --trigger-http --allow-unauthenticated
gcloud functions deploy playerstats --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=PlayerStats --trigger-http --allow-unauthenticated
gcloud functions deploy guildstats --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=GuildStats --trigger-http --allow-unauthenticated
gcloud functions deploy oauth2login --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=Oauth2Login --trigger-http --allow-unauthenticated
gcloud functions deploy oauth2callback --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=Oauth2Callback --trigger-http --allow-unauthenticated
gcloud functions deploy scanuserreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=ScanUserReports --trigger-http --allow-unauthenticated
gcloud functions deploy scanguildreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=ScanGuildReports --trigger-http --allow-unauthenticated
gcloud functions deploy scanrecentcharacterreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=ScanRecentCharacterReports --trigger-http --allow-unauthenticated

gcloud functions deploy coraideraccountclaim --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=CoraiderAccountClaim --retry --trigger-topic=coraideraccountclaim
gcloud functions deploy reportaccountclaim --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=ReportAccountClaim --retry --trigger-topic=reportaccountclaim
gcloud functions deploy fetchguildreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=FetchGuildReports --retry --trigger-topic=guildreports
gcloud functions deploy fetchreport2 --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=FetchReport --retry --trigger-topic=report
gcloud functions deploy updateplayerreport --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=UpdatePlayerReport --retry --trigger-topic=playerreport
gcloud functions deploy fetchuserreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=FetchUserReports --retry --trigger-topic=userreports
gcloud functions deploy fetchrecentcharacterreports --gen2 --runtime=go116 --region=europe-west2 --source=gs://raidlogscan_sources/source.zip --entry-point=FetchRecentCharacterReports --retry --trigger-topic=recentcharacterreports
