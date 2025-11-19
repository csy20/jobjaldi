package com.example.jobjaldi

import android.util.Log
import dev.csy.jobagent.jobagent.Jobagent
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {

    companion object {
        private const val CHANNEL_NAME = "jobagent"
        private const val EMPTY_JSON = "[]"
        private const val TAG = "JobagentChannel"
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        Jobagent.touch()

        MethodChannel(
            flutterEngine.dartExecutor.binaryMessenger,
            CHANNEL_NAME
        ).setMethodCallHandler { call, result ->
            when (call.method) {
                "scrapeMany" -> handleScrapeMany(call.arguments, result)
                else -> result.notImplemented()
            }
        }
    }

    private fun handleScrapeMany(arguments: Any?, result: MethodChannel.Result) {
        val cfg = arguments as? String
        if (cfg.isNullOrEmpty()) {
            Log.w(TAG, "scrapeMany expects a non-empty JSON string argument")
            result.success(EMPTY_JSON)
            return
        }

        try {
            val payload = Jobagent.scrapeMany(cfg)
            result.success(payload)
        } catch (e: Exception) {
            Log.e(TAG, "Jobagent.scrapeMany failed", e)
            result.success(EMPTY_JSON)
        }
    }
}
