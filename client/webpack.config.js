const path = require('path');

module.exports = {
  entry: './src/js/firebase-auth.js',
  output: {
    filename: 'firebase-auth.js',
    path: path.resolve(__dirname, 'public/js'),
    clean: false
  },
  mode: 'development',
  devtool: 'source-map',
  resolve: {
    extensions: ['.js']
  }
};