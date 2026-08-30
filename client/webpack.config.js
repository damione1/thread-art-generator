const path = require('path');

module.exports = (env, argv) => {
  const isProduction = argv.mode === 'production';

  return {
    entry: {
      main: './src/js/main.js',
      'art-upload': './src/ts/art-upload.ts',
      'status-poll': './src/ts/status-poll.ts',
      'password-auth': './src/ts/password-auth.ts',
    },
    output: {
      filename: '[name].js',
      chunkFilename: '[name].js',
      path: path.resolve(__dirname, 'public/js'),
      publicPath: '/static/js/',
      clean: false
    },
    mode: argv.mode || 'development',
    devtool: isProduction ? 'source-map' : 'eval-source-map',
    resolve: {
      extensions: ['.ts', '.js'],
      fallback: {
        fs: false,
        path: false,
        crypto: false,
        os: false,
      },
    },
    experiments: {
      asyncWebAssembly: true,
    },
    module: {
      rules: [
        {
          test: /\.ts$/,
          use: 'ts-loader',
          exclude: /node_modules/
        }
      ]
    },
    optimization: {
      splitChunks: {
        chunks: 'all',
        cacheGroups: {
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            chunks: 'initial',
          },
          rpc: {
            test: /[\\/]src[\\/](ts[\\/]rpc\.ts|gen[\\/])/,
            name: 'rpc',
            chunks: 'all',
            priority: 10,
          },
        },
      },
    }
  };
};
