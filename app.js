'use strict';
var log4js = require('log4js');
var logger = log4js.getLogger('PhantomClient');
var express = require('express');
var bodyParser = require('body-parser');
var app = express();
var https = require('https');
var fs = require('fs')
var cors = require('cors');

require('./config.js');
var hfc = require('fabric-client');
var invoke = require('./app/invoke-transaction.js');
var query = require('./app/query.js');

var host = process.env.HOST || hfc.getConfigSetting('host');
var port = process.env.PORT || hfc.getConfigSetting('httpsPort');

app.options('*', cors());
app.use(cors());
//support parsing of application/json type post data
app.use(bodyParser.json());
//support parsing of application/x-www-form-urlencoded post data
app.use(bodyParser.urlencoded({
	extended: false
}));

/**********************************  Start server ************** */

const server = https.createServer({
		key: fs.readFileSync('./artifacts/certs/serversk.pem'),
		cert: fs.readFileSync('./artifacts/certs/cert.pem')
	}, app)
	.listen(port, function () {
		console.log('*************** Server is listening on https://%s:%s  ******************', host, port);
	})
server.timeout = 240000;

function getErrorMessage(field) {
	var response = {
		success: false,
		message: field + ' field is missing or Invalid in the request'
	};
	return response;
}

///////////////////////// REST ENDPOINTS START HERE ///////////////////////////

// Invoke transaction on chaincode on target peers
app.post('/channels/:channelName/chaincodes/:chaincodeName', async function (req, res) {
	logger.debug('==================== INVOKE ON CHAINCODE ==================');
	try {

		var chaincodeName = req.params.chaincodeName;
		var channelName = req.params.channelName;
		var fcn = req.body.fcn;
		var args = req.body.args;
		logger.debug('channelName  : ' + channelName);
		logger.debug('chaincodeName : ' + chaincodeName);
		logger.debug('fcn  : ' + fcn);
		logger.debug('args  : ' + args);
		if (!chaincodeName) {
			res.json(getErrorMessage('\'chaincodeName\''));
			return;
		}
		if (!channelName) {
			res.json(getErrorMessage('\'channelName\''));
			return;
		}
		if (!fcn) {
			res.json(getErrorMessage('\'fcn\''));
			return;
		}
		if (!args) {
			res.json(getErrorMessage('\'args\''));
			return;
		}

		
		
		//let message = await invoke.invokeChaincode(channelName, chaincodeName, fcn, args);
		res.send("200");
	} catch (err) {
		logger.error('Http error::::::::', err);
		return res.status(500).send(err);
	}
});
// Query on chaincode on target peers
app.get('/channels/:channelName/chaincodes/:chaincodeName', async function (req, res) {
	logger.debug('==================== QUERY BY CHAINCODE ==================');
	var channelName = req.params.channelName;
	var chaincodeName = req.params.chaincodeName;
	let args = req.query.args;
	let fcn = req.query.fcn;
	let peer = req.query.peer;

	logger.debug('channelName : ' + channelName);
	logger.debug('chaincodeName : ' + chaincodeName);
	logger.debug('fcn : ' + fcn);
	logger.debug('args12 : ' + args);
	try {
		if (!chaincodeName) {
			res.json(getErrorMessage('\'chaincodeName\''));
			return;
		}
		if (!channelName) {
			res.json(getErrorMessage('\'channelName\''));
			return;
		}
		if (!fcn) {
			res.json(getErrorMessage('\'fcn\''));
			return;
		}
		if (!args) {
			res.json(getErrorMessage('\'args\''));
			return;
		}
		args = args.replace(/'/g, '"');
		args = JSON.parse(args);


		let message = await query.queryChaincode(peer, channelName, chaincodeName, args, fcn, req.username, req.orgname);
		res.send(message);
	} catch (err) {
		logger.error('Http error', err)
		return res.status(500).send()
	}
});
//  Query Get Block by BlockNumber
app.get('/channels/:channelName/blocks/:blockId', async function (req, res) {
	logger.debug('==================== GET BLOCK BY NUMBER ==================');
	let blockId = req.params.blockId;
	let peer = req.query.peer;
	logger.debug('channelName : ' + req.params.channelName);
	logger.debug('BlockID : ' + blockId);
	logger.debug('Peer : ' + peer);
	if (!blockId) {
		res.json(getErrorMessage('\'blockId\''));
		return;
	}

	let message = await query.getBlockByNumber(peer, req.params.channelName, blockId, "", "");
	res.send(message);
});
// Query Get Transaction by Transaction ID
app.get('/channels/:channelName/transactions/:trxnId', async function (req, res) {
	logger.debug('================ GET TRANSACTION BY TRANSACTION_ID ======================');
	logger.debug('channelName : ' + req.params.channelName);
	let trxnId = req.params.trxnId;
	let peer = req.query.peer;
	if (!trxnId) {
		res.json(getErrorMessage('\'trxnId\''));
		return;
	}

	let message = await query.getTransactionByID(peer, req.params.channelName, trxnId, req.username, req.orgname);
	res.send(message);
});

//Query for Channel Information
app.get('/channels/:channelName', async function (req, res) {
	logger.debug('================ GET CHANNEL INFORMATION ======================');
	logger.debug('channelName : ' + req.params.channelName);
	let peer = req.query.peer;

	let message = await query.getChainInfo(peer, req.params.channelName, req.username, req.orgname);
	res.send(message);
});
//Query for Channel instantiated chaincodes
app.get('/channels/:channelName/chaincodes', async function (req, res) {
	logger.debug('================ GET INSTANTIATED CHAINCODES ======================');
	logger.debug('channelName : ' + req.params.channelName);
	let peer = req.query.peer;

	let message = await query.getInstalledChaincodes(peer, req.params.channelName, 'instantiated', req.username, req.orgname);
	res.send(message);
});