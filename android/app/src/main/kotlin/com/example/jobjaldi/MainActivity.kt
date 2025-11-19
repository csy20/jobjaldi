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
            Log.d(TAG, "Calling scrapeMany with config: $cfg")
            val payload = Jobagent.scrapeMany(cfg)
            Log.d(TAG, "scrapeMany returned payload length: ${payload?.length ?: 0}")
            if (payload.isNullOrEmpty()) {
                Log.w(TAG, "scrapeMany returned empty payload")
                result.success(EMPTY_JSON)
            } else {
                result.success(payload)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Jobagent.scrapeMany failed", e)
            Log.e(TAG, "Error message: ${e.message}")
            Log.e(TAG, "Error stack trace: ${e.stackTraceToString()}")
            // Return error to Flutter instead of empty JSON
            result.error("SCRAPE_ERROR", "Failed to scrape jobs: ${e.message}", null)
        }
    }
}
