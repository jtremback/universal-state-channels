import express from 'express'
import bodyParser from 'body-parser'
import { Logic } from '../logic.js'

export default function (globals) {
  let app = express()
  app.use(bodyParser.json())
  
  const logic = new Logic(globals)

  const calls = [
    ['proposeChannel', 'propose_channel'],
    ['acceptChannel', 'accept_channel'],
    ['proposeUpdate', 'propose_update'],
    ['acceptUpdate', 'accept_update'],
    ['postUpdate', 'post_update'],
    ['startChallengePeriod', 'start_challenge_period']
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
