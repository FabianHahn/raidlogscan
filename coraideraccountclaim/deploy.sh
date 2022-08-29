#!/bin/bash
gcloud functions deploy coraideraccountclaim --gen2 --runtime=go116 --region=europe-west2 --source=. --entry-point=CoraiderAccountClaim --trigger-topic=coraideraccountclaim
