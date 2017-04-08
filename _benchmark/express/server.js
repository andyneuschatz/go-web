var compression = require('compression')
var express = require('express')
var app = express()

//app.use(compression);

// respond with "hello world" when a GET request is made to the homepage
app.get('/json', (req, res) => {
	res.send({ message: 'hello world!' })
})

var server = app.listen(9090, () => {
	console.log('Example app listening on port 9090!')
});