const path = require("path");

module.exports = {
	// mode: "production",
	entry: "./app.js",
	output: {
		path: path.resolve(__dirname, "public"),
		filename: "bundle.js"
	},
	/*
	optimization: {
		usedExports: true,
	},
	*/
	devtool: 'inline-source-map',
	module: {
		rules: [
			{
				test: /\.tsx?$/,
				use: 'ts-loader',
				exclude: /node_modules/,
			},
		],
	},
	resolve: {
		extensions: [ '.tsx', '.ts', '.js' ],
	},
};
