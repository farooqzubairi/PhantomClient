
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
async function getFabricClient () {
	let config = '-connection-profile-path';
	let client = hfc.loadFromConfig(hfc.getConfigSetting('network'+config));	
	await client.initCredentialStores();

	await client.setUserContext({
	username: "admin",
	password: "48378ab443"
});
	return client;
}


var getLogger = function(moduleName) {
	var logger = log4js.getLogger(moduleName);
	logger.setLevel('INFO');
	return logger;
};

exports.getFabricClient = getFabricClient;
exports.getLogger = getLogger;
