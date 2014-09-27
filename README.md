deploy.io
=========

Scale, Secure, Tune &amp; Monitor your web app effortlessly. Just Deploy.

Usage: deploy.js --alias ALIAS --key KEY --secret SECRET [OPTIONS]

Required:
 - alias:  Your consumer alias.
 - key:    Your oauth consumer key.
 - secret: Your oauth consumer secret token.

Note:
  alias, key and secret can also be read from your environment
  via exporting ALIAS, KEY, and/or SECRET with your credentials.

Examples:

$ ./deploy.js --alias ALIAS --key KEY --secret SECRET --name test123 --url http://www.example.com

$ ./deploy.js --alias ALIAS --key KEY --secret SECRET --url http://www.example.com
