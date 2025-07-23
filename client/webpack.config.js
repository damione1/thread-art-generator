const path = require('path');

module.exports = {
  entry: {
    main: './src/js/main.js',
    'firebase-auth': './src/js/firebase-auth.js'
  },
  output: {
    filename: '[name].js',
    path: path.resolve(__dirname, 'public/js'),
    clean: false
  },
  mode: 'development',
  devtool: 'source-map',
  resolve: {
    extensions: ['.js']
  },
  optimization: {
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          chunks: 'all',
        },
      },
    },
  }
};