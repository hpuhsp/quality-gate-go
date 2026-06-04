package com.example;

import android.app.Activity;
import android.os.Bundle;
import android.widget.TextView;

public class MainActivity extends Activity {

    // BUG: Hardcoded API key
    private static final String API_KEY = "AKIA1234567890ABCDEF";
    private static final String SECRET = "LTAI5tAbCdEfGhIjKlMnOpQrStUvWxYz";

    // BUG: Empty catch block
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        try {
            TextView tv = new TextView(this);
            tv.setText("Hello World");
            setContentView(tv);
        } catch (Exception e) {
            // silently ignore
        }
    }

    // BUG: SQL injection in content provider
    public void queryUser(String userName) {
        String selection = "name = '" + userName + "'";
        getContentResolver().query(
            android.net.Uri.parse("content://com.example.provider/users"),
            null, selection, null, null
        );
    }

    // BUG: Braces mismatch
    public void doSomething() {
        if (true) {
            for (int i = 0; i < 10; i++) {
                System.out.println(i);
        }
    }
}
