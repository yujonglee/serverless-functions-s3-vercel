# Serverless Functions
## Setup

```shell
yarn install && yarn setup
```

## Run(Local)

```shell
yarn start
```

It will use environment variables specified in Vercel.

## Deploy

```shell
yarn deploy
```

## ENV

```env
BUCKET_NAME=""
BUCKET_REGION=""
ACCESS_KEY_ID=""
SECRET_KEY=""
```

Note that `.env` file will not work.
You need to set environment variables in vercel, and use it with `os.Getenv`.

## Example
Object name in S3 is `${userId}.${fileName}`.

```js
const URL = 'something';

const res = await fetch(`${url}/api/presigned`, {
  headers: {
    Accept: 'application/json',
    'Content-Type': 'application/json',
  },
  method: 'POST',
  body: JSON.stringify([{ userId, fileName }]),
});

console.log(await res.json()); // ["some url for 1.txt"] (String[])
```
```js
const URL = 'something';

const res = await fetch(`${URL}/api/zipped`, {
  headers: {
    Accept: 'text/plain',
    'Content-Type': 'application/json',
  },
  method: 'POST',
  body: JSON.stringify([
    {
      "userId": "example1",
      "fileName": "1.txt"
    },
    {
      "userId": "example1",
      "fileName": "2.txt"
    },
  ]),
});

await res.text(); // "some url for zipped file of 1.txt and 2.txt" (String)
```
