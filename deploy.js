#!/usr/bin/env node
/***
 * Setup:
 *
 * npm install maxcdn subarg async
 *
 ***/
var MaxCDN, subarg, async;
try {
    MaxCDN = require('maxcdn');
    subarg = require('subarg');
    async  = require('async');
} catch (e) {
    console.error("Setup:");
    console.error(" ");
    console.error("$ npm install maxcdn subarg async ");
    console.error(" ");
    console.trace(e);
    process.exit(1);
}
var opts   = subarg(process.argv.slice(2));

if (opts.help || opts.h) usage();
 
[ 'alias', 'key', 'secret', 'type', 'name', 'url' ].forEach(function(opt) {
    // support ALIAS, KEY, SECRET from environment
    opts[opt] = opts[opt] || process.env[opt.toUpperCase()];
 
    // ensure required params exist
    if (typeof opts[opt] === 'undefined') {
        usage(1, "Missing required argument: "+opt); // usage with error status
    }
    
    if (opts[opt] === '') {
        usage(1, "Argument can't be an empty string: "+opt); // usage with error status
    }
});

// init MaxCDN
var maxcdn = new MaxCDN(opts.alias, opts.key, opts.secret);

createZone(opts.type, opts.name, opts.url);

function createZone(type, name, url) {
    process.stdout.write('creating ' + type + ' zone ' + name + '... ');
    maxcdn.post('zones/'+type+'.json', {name: name, url: url}, function (err, res) {
        if (res == undefined) {
            console.log('zone creation failed');
        } else {
            console.log('complete', res);
        }
    });
}
 
function usage(status, error) {
    status = status || 0; // default to zero exit status
 
    if (error) {
        console.error("ERROR:", error);
        console.error(" ");
    }
 
    console.log("Usage: deploy.js --alias ALIAS --key KEY --secret SECRET [OPTIONS]");
    console.log(" ");
    console.log(" Required:");
    console.log(" - alias:  Your consumer alias.");
    console.log(" - key:    Your oauth consumer key.");
    console.log(" - secret: Your oauth consumer secret token.");
    console.log(" ");
    console.log(" Note:");
    console.log("   alias, key and secret can also be read from your environment");
    console.log("   via exporting ALIAS, KEY, and/or SECRET with your credentials.");
    console.log(" ");
    console.log(" Examples:");
    console.log(" ");
    console.log(" $ ./deploy.js --alias ALIAS --key KEY --SECRET --type pull --name test123 --url http://www.example.com");
    console.log(" ");
    process.exit(status);
}