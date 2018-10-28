var util = require('util');
var path = require('path');
var hfc = require('fabric-client');

var file = 'network-config%s.json';

var env = process.env.TARGET_NETWORK;
if (env)
	file = util.format(file, '-' + env);
else
	file = util.format(file, '');
hfc.setConfigSetting('network-connection-profile-path',path.join(__dirname, 'artifacts' ,file));
hfc.setConfigSetting('org-connection-profile-path',path.join(__dirname, 'artifacts', 'org-config.json'));

// some other settings the application might need to know
hfc.addConfigFile(path.join(__dirname, 'config.json'));
