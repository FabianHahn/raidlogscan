# raidlogscan

[Live instance link](https://accountstats-ic7qdt5tla-nw.a.run.app/?account_name=Khumba)

## Features
 * Parses raidlogs from warcraftlogs.com and generates leaderboards of who everyone played with the most.
 * Allows grouping multiple characters per human player, across servers too.
 * Can scan all raids published under a guild / raid team on Warcraftlogs.
 * Allows logging into a personal Warcraft Logs Account using oauth2, and then scanning personal logs as well as recent character logs.
 * Fully deployed as Cloud Functions to Google Cloud, making it very cheap to run.
 * Using Firebase/Datastore as database, and Pub/Sub for events and triggers.
 * Written in Go 1.16.
 * MIT license.
 
## Implementation

### Data model

A **report** entity stores the details for a single scanned raid report, including all of its players that participated.

A **player** entity stores the details for a player character that appeared in at least one report.
It also stores all the reports it appeared in, all the other players ("coraiders") and the number of times ("count") it raided with them in those reports.
Further, it also stores mappings from coraider player IDs to account names that group them.

### Data flow

A list of report codes to scan is generated in one of three ways:
 * An input guild ID of a raid team to be scanned.
 * An input user ID of a Warcraftlogs account for which public personal logs should be scanned.
 * An input character ID for a list of recent reports to be scanned.

For each report code, the list of players in that report is fetched and stored in a report entity in the database.
For each participating player, an event is emitted to update (or create) the respective player entity for this report.
This will:
 * Insert the report into the list of reports the player participated in.
 * Update the coraiders and their appearance counts.
 * If the player is claimed by an account name, send "coraider account claim" events to all newly appeared coraiders.

A coraider account claim event then results in the targeted player entity's mapping from known coraider player IDs to account names to be updated.
This denormalization allows us to fetch details for account names on a per player basis without having to store separate entities for accounts themselves.

### Identifiers

| Id name | Description |
| ------- | ----------- |
| Report code | Parsed raid report on warcraftlogs.com, i.e. the alphanumeric identifier bit in the log URL |
| Player ID | Warcraftlogs ID of a player in a report, i.e. a WoW character that has participated in a raid |
| Account name | Grouping of multiple player IDs to denote that they belong to the same human player, given as user input through the web UI |
| Guild ID | Warcraftlogs ID of a guild or raid team for which reports can be fetched, i.e. what shows up as "Calender" in the Warcraftlogs UI |
| User ID | Warcraftlogs ID of a user account on the site, i.e. what is used to upload reports. |
| Character ID | Warcraftlogs ID of a character claimed by a personal account in the Warcraftlogs UI, confusingly not the same as a player ID. Used to fetch recent logs for that character. |
