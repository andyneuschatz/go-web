var express = require('express')
var app = express()

// respond with "hello world" when a GET request is made to the homepage
app.get('/json', function (req, res) {
  setTimeout(function () {
    res.send({ message: 'hello world!' })
  }, 1000)
})

app.listen(9090, function () {
  console.log('Example app listening on port 9090!')
})