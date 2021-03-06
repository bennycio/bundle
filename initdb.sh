#!/bin/sh
mongo --eval 'db.plugins.createIndex( { name: 1 }, { unique: true, collation: { locale: "en", strength: 1 } } )' plugins
mongo --eval 'db.plugins.createIndex( { name: "text", description: "text" } )' plugins
mongo --eval 'db.users.createIndex( { username: 1 }, { unique: true, collation: { locale: "en", strength: 1 } } )' users
mongo --eval 'db.users.createIndex( { email: 1 }, { unique: true, collation: { locale: "en", strength: 1 } } )' users