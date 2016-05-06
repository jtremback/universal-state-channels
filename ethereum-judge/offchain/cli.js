let argv = require('minimist')(process.argv.slice(2))
if (commands[argv._[0]]) {
  commands[argv._[0]](argv)
} else if (argv._[0]) {
  console.log(argv._[0] + ' is not a command')
} else {
  console.log('please enter a command')
}