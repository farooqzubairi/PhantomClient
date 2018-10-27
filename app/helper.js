
'use strict';
var log4js = require('log4js');
var logger = log4js.getLogger('helper');
logger.setLevel('INFO');

var path = require('path');
var util = require('util');
var copService = require('fabric-ca-client');

var hfc = require('fabric-client');
hfc.setLogger(logger);
var ORGS = hfc.getConfigSetting('network-config');

var clients = {};
var channels = {};
var caClients = {};

var sleep = async function (sleep_time_ms) {
	return new Promise(resolve => setTimeout(resolve, sleep_time_ms));
}

async function getFabricClient (userorg, username) {
	logger.info('getFabricClient - ****** START %s %s', userorg, username)
	// get a fabric client loaded with a connection profile for this org
	let config = '-connection-profile-path';
	logger.info(":::::::"+hfc.getConfigSetting('network'+config));
	// build a client context and load it with a connection profile
	// lets only load the network settings and save the client for later
	let client = hfc.loadFromConfig(hfc.getConfigSetting('network'+config));
	
	
	await client.initCredentialStores();

	await client.setUserContext({
	username: "admin",
	password: "48378ab443"
});
	/*	
	if(username) {

		
		let user = await client.getUserContext(username, true);
		if(!user) {
			throw new Error(util.format('User was not found :', username));
		} else {
			logger.info('User %s was found to be registered and enrolled', username);
		}
	}
	logger.debug('getFabricClient - ****** END %s %s \n\n', userorg, username)
*/
	return client;
}


var getLogger = function(moduleName) {
	var logger = log4js.getLogger(moduleName);
	logger.setLevel('INFO');
	return logger;
};

exports.getClientForOrg = getFabricClient;
exports.getLogger = getLogger;
