duitmail - simple, secure mail client

this is work in progress.
this doesn't actually work yet, actually reading email isn't implemented other
than reading mail files at local/mailboxes/*/*. copy some of your mail files
there before starting.

# instructions

multiple fonts are read from $font, $boldfont and $fontawesome, eg:

	export font=/mnt/font/Lato-Regular/15a/font
	export boldfont=/mnt/font/Lato-Bold/15a/font
	export fontawesome=/mnt/font/FontAwesome5FreeSolid/12a/font


# goals

- imap4 support for reading
- smtp for sending
- multiple accounts

# todo

- dynamically fetch emails using imap, with local cache of emails
- better error handling (don't quit on error)
- properly content-transfer-encode outgoing messages
- send emails with attachments
- add support for multiple accounts
