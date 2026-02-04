# Summary

The program is called secled.

Secled is a password manager. It is a command line tool that helps with kubectl and curl commands if user needs to use token, or webhook secret or anything else that couldn't be stored in the text files and can't be copied in the commands (otherwise you can see them in the history or they make their way up to some git repo (very bad!)).

Secled stores data in key-value pairs. The values are encrypted and stored in a binary file. In basic setup user will create ~/secled directory, clonse the repo, builds it and runs the binary from that dir. All files are in ~/secled/secled/bin. User will add alias secled='~/secled/secled/bin/secled' into .bashrc and runs like user@box:~$ secled. The storage file is ledger.encrypted in the same dir ~/secled/secled/bin.

Secled has explicit login and logout. The command "secled login" asks for the master password. If it's the first time, this will be the master password of the ledger file and secled states that fact (also creates first entry with key "initial" that holds current date and minimal information about the system). Secled prints a shell snippet that sets SECLED_MASTER, and the user runs it with eval in their shell. SECLED_MASTER is used whenever "secled add :key:" or "secled get :key:" is called to decrypt stored data of that key.

When user wants to add a key/data into the ledger using "secled add 'my-long-key with-space'" that opens standard input (like sudo asks password) where user can paste the key. When hitting enter secled encrypts the data. This requires that SECLED_MASTER is set. If not, return 1 with an error that login is required. Secled stores the key-data pair in an encrypted way in the same directory where the binary of secled locates.

When user wants to retrieve the key/data from ledger, user must run "secled get 'my-long-key with-space'" that prints out the decrypted data of the key. User most likely will use it in pipeline of commands. This command requires env parameter SECLED_MASTER is set. If not, return 1 "Error: login required (SECLED_MASTER is not set)".

If user forgets what keys are stored "secled list" could be used to list all the keys, but not the data. This must work even without password.

# Objective 

Secled helps with tokens, keys and password so the user does not have to copy paste them and also has persistant storage. The storage is a binary file. The keys are not encrypted and are searchable but the data (passwords, tokens, etc) must be encrypted with strong symmetrical way. Secled encourages users to use at least 8 char password with some extra chars. This is not required but recommended.

# Usecases

## Initial run

### Pre conditions
1. Go is installed in the system
2. mkdir ~/secled
3. cd ~/secled
4. git clone git@github.com:triiberg/secled.git
5. cd secled
6. go build -o bin/secled ./cmd
7. build is successful ~/secled/secled/bin/secled file exists and is executable

### Using secled first time
1. echo 'alias secled="~/secled/secled/bin/secled"' >> ~/.bashrc # if its a linux, it has to work with win and mac too
2. echo 'alias secled-login="eval \"$(secled login)\""' >> ~/.bashrc
3. run "secled-login", it will ask the password
4. enter the password
5. verify file ~/secled/secled/bin/, there must be a file called ledger.encrypted
6. list the keys (there must be the example key "initial"): secled list
7. get the data of the key "initial" by running $ secled get initial, that must print out the date when it was created and basic system info
8. see if the SECLED_MASTER is set by calling echo $SECLED_MASTER

# Project

## Technical details

### Stack

- Go
- must be compilable and runnable on modern windows, linux and mac
- use minimal set of libraries: the less dependencies the less security issues

## Functionality

### Commands
- secled login: prompts for master password, prints a shell snippet that sets SECLED_MASTER
- secled logout: prints a shell snippet that unsets SECLED_MASTER
- secled list: displays all keys that are stored in the ledger
- secled add <key>: will ask what is the data of the key using stdin, encrypts the data and stores in the file
- secled get <key>: using SECLED_MASTER password decrypts data of the key and prints out (so it would be easy to use in like kubectl create secret generic my-secret --from-literal=key1=`secled get ghcr-password` ...)
- secled update <key>: replaces data of existing key, requires SECLED_MASTER
- secled remove <key>: deletes a key, requires SECLED_MASTER
- secled generate-uuid <key>: generates a UUID v4 and stores it under key
- secled generate-64hex <key>: generates 64 hex chars (32 random bytes) and stores it under key

### Key rules
- the key is a single argument
- if it has spaces, the user must quote it in the shell, for example: secled get 'my key'
- add/generate must fail if the key already exists

### Ledger location
- ledger.encrypted is always stored in the same directory as the secled binary

### Initial entry
- key name: initial
- value fields: created_at (RFC3339), hostname, goos, goarch
- must use Go standard library only for this info (no OS-specific libraries)

## Not needed functionality

- no swiping the data
- no API's 
- no backups
- no database (sqllight, duck etc)

## Programming tips

- make it simple and understandable for humans
- do not use interfaces if not needed
- divide the code into few files where the name specifies the part of the project
- less files means its easier to audit by humans
- when creating functions, keep in mind you have to write tests
- write tests and keep code coverage up like 80% (only main.go is allowed not to have tests)
- tests next to the files

