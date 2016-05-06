import express from 'express'
import bodyParser from 'body-parser'
import { Logic } from '../logic.js'

export default function (globals) {
  let app = express()
  app.use(bodyParser.json())
  
  const logic = new Logic(globals)

  const calls = [
    ['addProposedChannel', 'add_proposed_channel'],
    ['addProposedUpdate', 'add_proposed_update'],
    ['addAcceptedUpdate', 'add_accepted_update']
  ]

  for (let call of calls) {
    app.post(call[0], function (req, res) {
      logic[call[1]](req.body)
      .then(result => {
        res.send(result)
      })
      .catch(error => {
        status(500).send({ error })
      })
    })
  }

  return app
}
