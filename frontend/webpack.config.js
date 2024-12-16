const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const webpack = require('webpack');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;
const TerserPlugin = require('terser-webpack-plugin');

module.exports = {
  mode: 'production',
  entry: './src/index.tsx',
  output: {
    path: path.resolve(__dirname, 'build'),
    filename: '[name].[contenthash].js',
    clean: true,
    publicPath: '/'
  },
  devtool: 'source-map',
  module: {
    rules: [
      {
        test: /\.(ts|tsx)$/,
        use: [
          {
            loader: 'ts-loader',
            options: {
              transpileOnly: true,
              compilerOptions: {
                noEmit: false,
                skipLibCheck: true
              }
            }
          }
        ],
        exclude: /node_modules/,
      },
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader'],
      },
    ],
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.js', '.jsx'],
    alias: {
      'react': path.resolve('./node_modules/react'),
      'react-dom': path.resolve('./node_modules/react-dom'),
      'styled-components': path.resolve('./node_modules/styled-components')
    },
    fallback: {
      assert: require.resolve('assert/'),
      buffer: require.resolve('buffer/'),
      stream: require.resolve('stream-browserify'),
      util: require.resolve('util/'),
      'process': require.resolve('process/browser'),
      'path': require.resolve('path-browserify'),
      'fs': false,
      'crypto': false
    }
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: './public/index.html',
      inject: true,
      minify: {
        removeComments: true,
        collapseWhitespace: true,
        removeRedundantAttributes: true,
        useShortDoctype: true,
        removeEmptyAttributes: true,
        removeStyleLinkTypeAttributes: true,
        keepClosingSlash: true,
        minifyJS: true,
        minifyCSS: true,
        minifyURLs: true,
      }
    }),
    new webpack.DefinePlugin({
      'process.env': {
        NODE_ENV: JSON.stringify('production'),
        REACT_APP_API_KEY: JSON.stringify(process.env.REACT_APP_API_KEY),
      },
      'process.browser': true
    }),
    new webpack.ProvidePlugin({
      process: 'process/browser',
      Buffer: ['buffer', 'Buffer'],
      React: 'react'
    }),
    new webpack.optimize.ModuleConcatenationPlugin(),
    new BundleAnalyzerPlugin({
      analyzerMode: 'static',
      openAnalyzer: false,
      reportFilename: 'bundle-report.html'
    })
  ],
  optimization: {
    minimize: true,
    minimizer: [
      new TerserPlugin({
        terserOptions: {
          parse: {
            ecma: 8,
          },
          compress: {
            ecma: 5,
            warnings: false,
            comparisons: false,
            inline: 2,
          },
          mangle: {
            safari10: true,
          },
          output: {
            ecma: 5,
            comments: false,
            ascii_only: true,
          },
        },
        parallel: true,
      }),
    ],
    splitChunks: {
      chunks: 'all',
      name: false,
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          chunks: 'all'
        }
      }
    },
    runtimeChunk: {
      name: entrypoint => `runtime-${entrypoint.name}`
    }
  },
  performance: {
    maxEntrypointSize: 512000,
    maxAssetSize: 512000,
    hints: 'warning'
  }
}; 