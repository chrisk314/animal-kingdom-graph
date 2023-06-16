const express = require("express");
const cors = require("cors");
const app = express();

// Configure CORS
var corsOptions = {
  origin: ['http://localhost:5173', 'http://orion:5173',],
  optionsSuccessStatus: 200, // some legacy browsers (IE11, various SmartTVs) choke on 204
}
app.use(cors(corsOptions));

app.listen(5173, () => {
  console.log("Application started and Listening on port 3000");
});

// serve static files
app.use(express.static(__dirname));

app.get("/", (req, res) => {
  res.sendFile(__dirname + "/index.html");
});