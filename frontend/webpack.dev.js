const { merge } = require('webpack-merge');
const common = require('./webpack.common.js');
const path = require('path');
const webpack = require('webpack');

module.exports = merge(common, {
  mode: 'development',
  devtool: 'inline-source-map',
  output: {
    path: path.resolve(__dirname, 'build'),
    filename: '[name].bundle.js',
    publicPath: '/'
  },
  plugins: [
    new webpack.DefinePlugin({
      'process.env': {
        NODE_ENV: JSON.stringify('development'),
        REACT_APP_API_KEY: JSON.stringify(process.env.REACT_APP_API_KEY),
      }
    })
  ],
  devServer: {
    static: [
      {
        directory: path.join(__dirname, 'build'),
      },
      {
        directory: path.join(__dirname, 'public'),
      }
    ],
    compress: true,
    port: 3000,
    hot: true,
    host: '0.0.0.0',
    allowedHosts: 'all',
    historyApiFallback: {
      index: '/index.html',
      rewrites: [
        { from: /.*/, to: '/index.html' }
      ]
    },
    client: {
      webSocketURL: 'ws://localhost/ws',
      progress: true,
      logging: 'verbose'
    },
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
    watchFiles: {
      paths: ['src/**/*', 'public/**/*'],
      options: {
        usePolling: true,
        poll: 1000,
      },
    },
    devMiddleware: {
      writeToDisk: true,
      stats: 'detailed'
    },
  }
}); 