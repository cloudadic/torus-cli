'use strict';

var Promise = require('es6-promise').Promise;

var credentials = require('./credentials');
var cValue = require('./value');
var harvest = require('./harvest');

var set = exports;
set.output = {};

set.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (ctx.params.length < 2) {
      return reject(new Error('You must provide two parameters'));
    }

    var value = cValue.create(ctx.params[1]);
    var params = harvest(ctx);

    return credentials.create(ctx.session, params, value)
      .then(resolve).catch(reject);
  });
};

set.output.success = function () {
  console.log('Credential has been set!');
};

set.output.failure = function () {
  console.log('It failed ;(');
};