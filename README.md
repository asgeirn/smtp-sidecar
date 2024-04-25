# SMTP Sidecar

This Go project is intended to be used as a sidecar container for applications which need to send e-mails.

The idea is to expose a simple SMTP server that then uses the GMail API to send the email, removing
the need for your application to support the GMail API, or the need to set up authenticated SMTP 
accounts for the application to use.

Furthermore, the code can be modified to add rules and checks to avoid spamming.
