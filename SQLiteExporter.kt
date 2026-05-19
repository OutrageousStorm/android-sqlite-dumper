// SQLiteExporter.kt -- Android Kotlin utility to export SQLite databases
// Usage: In your Activity, call ExportDatabase(context, "com.example.app", "app.db")

package com.outrageousstorm.sqlite

import android.content.Context
import android.database.Cursor
import android.database.sqlite.SQLiteDatabase
import java.io.File
import java.io.FileOutputStream
import org.json.JSONArray
import org.json.JSONObject

object SQLiteExporter {
    fun exportToJson(context: Context, packageName: String, dbName: String): String {
        val dbFile = context.getDatabasePath(dbName)
        if (!dbFile.exists()) {
            return "{\"error\": \"Database not found\"}"
        }

        val db = SQLiteDatabase.openDatabase(dbFile.absolutePath, null, SQLiteDatabase.OPEN_READONLY)
        val result = JSONObject()

        // Get all table names
        val cursor: Cursor = db.rawQuery(
            "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'",
            null
        )

        while (cursor.moveToNext()) {
            val tableName = cursor.getString(0)
            val rows = JSONArray()

            // Get all rows from table
            val tableCursor = db.query(tableName, null, null, null, null, null, null)
            val columnNames = tableCursor.columnNames

            while (tableCursor.moveToNext()) {
                val row = JSONObject()
                for (i in columnNames.indices) {
                    row.put(columnNames[i], tableCursor.getString(i) ?: "NULL")
                }
                rows.put(row)
            }
            tableCursor.close()

            result.put(tableName, rows)
        }
        cursor.close()
        db.close()

        return result.toString(2)
    }

    fun exportToCSV(context: Context, packageName: String, dbName: String, table: String): String {
        val dbFile = context.getDatabasePath(dbName)
        val db = SQLiteDatabase.openDatabase(dbFile.absolutePath, null, SQLiteDatabase.OPEN_READONLY)

        val cursor = db.query(table, null, null, null, null, null, null)
        val columnNames = cursor.columnNames
        val csv = StringBuilder()

        // Header
        csv.append(columnNames.joinToString(",") + "\n")

        // Rows
        while (cursor.moveToNext()) {
            val row = mutableListOf<String>()
            for (i in columnNames.indices) {
                val val_str = cursor.getString(i) ?: ""
                row.add("\"${val_str.replace("\"", "\\\"")}\"")
            }
            csv.append(row.joinToString(",") + "\n")
        }

        cursor.close()
        db.close()

        return csv.toString()
    }

    fun backupDatabase(context: Context, dbName: String, outputPath: String): Boolean {
        return try {
            val source = context.getDatabasePath(dbName)
            val dest = File(outputPath)
            source.copyTo(dest, overwrite = true)
            true
        } catch (e: Exception) {
            false
        }
    }
}
