# How to use it

This daemon is ment to be used behind a Nginx proxy.

Using it on it's own is dangerous and not recommended!

## Sample nginx config

```
proxy_cache_path /var/lib/cache/mini levels=1:2 keys_zone=minify:10m max_size=10g
                 inactive=60m use_temp_path=off;

server {
    listen       80;
    server_name example.com;
    root /home/example/public_html;


location ~* \.(js|css)$ {
        proxy_cache minify;
        add_header X-Proxy-Cache $upstream_cache_status;
        proxy_cache_valid any 1h;
        proxy_cache_min_uses 1;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header Host $host;
        proxy_set_header X-Docroot $document_root;
        proxy_pass_request_headers      on;
        proxy_buffering        on;
        proxy_cache_valid      200  1d;
        proxy_pass http://127.0.0.1:3000;
        expires 1y;
    }

}
```
## Running the daemon

```./minifier -port=127.0.0.1:3000```