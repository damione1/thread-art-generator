const path = require('path');

module.exports = (env, argv) => {
  const isProduction = argv.mode === 'production';
  
  return {
    entry: {
      main: './src/ts/main.ts',
      'firebase-auth': './src/ts/firebase-auth.ts',
      'art-upload': './src/ts/art-upload.ts'
    },
    output: {
      filename: '[name].js',
      path: path.resolve(__dirname, 'public/js'),
      clean: false
    },
    mode: argv.mode || 'development',
    devtool: isProduction ? 'source-map' : 'eval-source-map',
    resolve: {
      extensions: ['.ts', '.js']
    },
    module: {
      rules: [
        {
          test: /\.ts$/,
          use: 'ts-loader',
          exclude: /node_modules/
        },
        {
          test: /firebase-auth\.(js|ts)$/,
          // Ensure firebase-auth is never tree-shaken by marking it as having side effects
          sideEffects: true
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
            chunks: 'all',
          },
        },
      },
      // Disable tree shaking in production for entry modules to prevent issues
      usedExports: !isProduction,
      sideEffects: false
    }
  };
};