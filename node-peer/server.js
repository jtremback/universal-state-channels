import express from 'express'
var app = express()

app.post('/propose_channel', function (req, res) {
  res.send('Hello World!')
})

app.post('/propose_update', function (req, res) {
  res.send('Hello World!')
})

app.post('/accept_update', function (req, res) {
  res.send('Hello World!')
})

app.listen(3000, function () {
  console.log('Example app listening on port 3000!')
})