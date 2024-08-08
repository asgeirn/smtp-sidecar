# SMTP Sidecar

This Go project is intended to be used as a sidecar container for applications which need to send e-mails.

The idea is to expose a simple SMTP server that then uses the GMail API to send the email, removing
the need for your application to support the GMail API, or the need to set up authenticated SMTP
accounts for the application to use.

Furthermore, the code can be modified to add rules and checks to avoid spamming.

## Configuration

The application / Docker container is configured using environment variables:

| Variable           | Description                                                      |
|--------------------|------------------------------------------------------------------|
| `SMTP_LISTEN`      | Address to listen to, defaults to `:2525`                        |
| `CREDENTIALS_JSON` | Path to the GMail API OAuth credentials file.                    |
| `TOKEN_JSON`       | Path to the access token JSON.                                   |
| `SENDERS`          | Comma-separated list of permitted senders (empty permits all)    |
| `RECIPIENTS`       | Comma-separated list of permitted recipients (empty permits all) |

**NOTE!** If both  `SENDERS` and `RECIPIENTS` are empty, all email addresses are permitted.
Do not combine this with listening on `:25` as all machines on your local network now can
act as spam robots!

## Initial setup of GMail API

Follow the instructions at https://developers.google.com/gmail/api/quickstart/go to enable the API
and configure OAuth credentials.

The application requires the `https://www.googleapis.com/auth/gmail.send` scope.

## First time login

When logging in for the first time, the application outputs an URL that you need to open
to grant access.  It also spins up a local web server to capture the token response on a
random port.  This can be tricky to do in a Docker container.  To work around this, run
the application once outside the container and save the `token.json` file.
