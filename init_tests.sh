#!/bin/bash
rm privkey
rm publickey

openssl ecparam -genkey -name secp521r1 -noout -out privkey 
openssl ec -in privkey -pubout -out publickey

cd ./utils      && ln ../.env .env; ln ../publickey publickey; cd ..
cd ./middleware && ln ../.env .env; ln ../publickey publickey; cd ..
cd ./routes     && ln ../.env .env; ln ../publickey publickey; cd ..
cd ./config     && ln ../.env .env; ln ../publickey publickey; cd ..
cd ./model      && ln ../.env .env; ln ../publickey publickey; cd ..
cd ./security   && ln ../.env .env; ln ../publickey publickey; ln ../privkey privkey; cd ..

true

