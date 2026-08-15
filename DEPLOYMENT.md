# 閮ㄧ讲鏋舵瀯涓庢惌寤烘祦绋嬶紙remote-wakeup锛?
> 鏈枃妗ｈ褰曟湰椤圭洰浠庨浂鎼缓鐨勫畬鏁存祦绋嬩笌褰撳墠鐢熶骇鏋舵瀯锛屼緵缁存姢鑰呭弬鑰冦€?> 鏁忔劅淇℃伅涓€寰嬩娇鐢ㄥ崰浣嶇锛屼笉鍖呭惈鐪熷疄 token / 瀵嗙爜銆?
## 1. 鏁翠綋鏋舵瀯

```
寮€鍙戣€呮湰鍦?(Windows)
   鈹? 鍐欎唬鐮?鈫?go vet / go test锛堟湰鍦拌嚜娴嬶級
   鈹? git commit + push
   鈻?GitHub (remote-wakeup 浠撳簱)
   鈹? 瑙﹀彂 GitHub Actions锛坵orkflow: deploy.yml锛?   鈻?CI/CD锛圙itHub 浜戠 runner锛寀buntu-latest锛?   鈹? 鈶?go vet ./... + go test ./...锛堣川閲忛棬绂侊級
   鈹? 鈶?浜ゅ弶缂栬瘧 arm64 浜岃繘鍒讹紙CGO_ENABLED=0 GOOS=linux GOARCH=arm64锛?   鈹? 鈶?閫氳繃 Cloudflare Tunnel (ssh.<your-domain>) SSH 鍒版澘瀛?   鈻?鏉垮瓙锛圤range Pi Zero 3锛?   鈹? 鍋滄湇鍔?鈫?scp 浜岃繘鍒?鈫?閲嶅惎
   鈻?/opt/wol-web/wol-web  锛坰ystemd: wol-web锛岀洃鍚?:5000锛?```

## 2. 鐢熶骇閮ㄧ讲浣嶇疆

| 椤?| 鍊?|
|---|---|
| 閮ㄧ讲鐩綍 | `/opt/wol-web/` |
| 浜岃繘鍒?| `/opt/wol-web/wol-web` |
| 閰嶇疆 | `/opt/wol-web/.env`锛堟潈闄?600锛?|
| systemd 鏈嶅姟 | `wol-web.service` |
| 鐩戝惉绔彛 | `5000` |
| 鍏綉鍏ュ彛 | `remote-wakeup.<your-domain>`锛圕loudflare 闅ч亾 鈫?http://localhost:5000锛?|
| CI 閮ㄧ讲鍏ュ彛 | `ssh.<your-domain>`锛圕loudflare 闅ч亾 鈫?ssh://localhost:22锛?|

## 3. 鏉垮瓙渚х粍浠?
| 缁勪欢 | 璇存槑 | 鐘舵€?|
|---|---|---|
| `cloudflared` | Cloudflare 闅ч亾瀹㈡埛绔紙token 妯″紡锛?| systemd 鏈嶅姟锛屽父椹?|
| `/opt/wol-web/` | 鐢熶骇閮ㄧ讲鐩綍锛坮oot 灞炰富锛屼簩杩涘埗 750 / .env 600锛?| 鐢?CI 鏇存柊 |
| `wol-web.service` | systemd 鍗曞厓锛屾寚鍚?`/opt/wol-web/wol-web` | 甯搁┗ |

## 4. Cloudflare 渚ч厤缃?
| 椤?| 閰嶇疆 | 鐢ㄩ€?|
|---|---|---|
| 鍩熷悕 | `<your-domain>` | 鎵樼鍦?Cloudflare |
| 闅ч亾 | token 妯″紡锛坱unnel run --token锛?| 鏉垮瓙涓诲姩杩?Cloudflare |
| Public Hostname | `remote-wakeup.<your-domain>` 鈫?`http://localhost:5000` | 鍏綉璁块棶 Web |
| Public Hostname | `ssh.<your-domain>` 鈫?`ssh://localhost:22` | CI/CD 閮ㄧ讲 SSH 鍏ュ彛 |
| Service Token | 鍦?Access 鈫?Service Auth 鍒涘缓 | CI 璁よ瘉锛堟牸寮?ID:SECRET锛?|
| Access 绛栫暐 | `ci-deploy`锛圫ervice Auth + token锛?| 鏀捐 CI 鐨?cloudflared 杩炴帴 |

## 5. GitHub 渚ч厤缃?
| 椤?| 鍊?| 璇存槑 |
|---|---|---|
| 浠撳簱 | `632-8nm/remote-wakeup` | 鈥?|
| Workflow | `.github/workflows/deploy.yml` | push main 瑙﹀彂 |
| Workflow | `.github/workflows/release.yml` | 鎵?v* tag 瑙﹀彂锛屽彂甯?Release |
| Secret: `BOARD_SSH_KEY` | 鏉垮瓙 `~/.ssh/deploy` 绉侀挜 | 浜戠 SSH 鐧诲綍鏉垮瓙 |
| Secret: `CLOUDFLARED_TOKEN` | `ClientID:ClientSecret` | cloudflared access 璁よ瘉 |

## 6. 浠庨浂鎼缓姝ラ

### 6.1 鏉垮瓙鍑嗗
```bash
# 瀹夎 cloudflared
curl -L --output /usr/local/bin/cloudflared \
  https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64
chmod +x /usr/local/bin/cloudflared

# 閰嶇疆鍏嶅瘑 sudo锛堜粎 systemctl/journalctl/tee锛屼緵 CI 浣跨敤锛?sudo tee /etc/sudoers.d/orangepi-systemd <<'EOF'
orangepi ALL=(ALL) NOPASSWD: /usr/bin/systemctl, /bin/systemctl, /usr/bin/journalctl, /usr/bin/tee
EOF
sudo chmod 440 /etc/sudoers.d/orangepi-systemd

# 鐢熸垚 CI 閮ㄧ讲瀵嗛挜
ssh-keygen -t ed25519 -N "" -f ~/.ssh/deploy
cat ~/.ssh/deploy.pub >> ~/.ssh/authorized_keys

# 鍒涘缓鐢熶骇閮ㄧ讲鐩綍
sudo mkdir -p /opt/wol-web
sudo chown orangepi:orangepi /opt/wol-web
```

### 6.2 Cloudflare 閰嶇疆
1. 鍩熷悕鎵樼鍦?Cloudflare锛坄<your-domain>`锛?2. 鏉垮瓙瀹夎 cloudflared锛岀敤 token 鎺ュ叆闅ч亾
3. 鍔?Public Hostname锛?   - `remote-wakeup.<your-domain>` 鈫?`http://localhost:5000`
   - `ssh.<your-domain>` 鈫?`ssh://localhost:22`
4. Access 鈫?Service Auth 鍒涘缓 Service Token锛堣涓?Client ID/Secret锛?5. Access 鈫?涓?`ssh.<your-domain>` 閰嶇疆 `ci-deploy` 绛栫暐锛圫ervice Auth锛?
### 6.3 GitHub 閰嶇疆
1. 浠撳簱鍔犱袱涓?Secret锛?   - `BOARD_SSH_KEY` = 鏉垮瓙 `~/.ssh/deploy` 绉侀挜鍏ㄦ枃
   - `CLOUDFLARED_TOKEN` = `ClientID:ClientSecret`
2. 鎺ㄩ€?`.github/workflows/deploy.yml`锛坧ush main 鑷姩閮ㄧ讲锛?3. 鎺ㄩ€?`.github/workflows/release.yml`锛堟墦 tag 鑷姩鏋勫缓 Release锛?
### 6.4 楠岃瘉
```bash
# 鍦ㄦ澘瀛愭祴璇曢毀閬?SSH锛堟ā鎷?CI锛?cloudflared access tcp --hostname ssh.<your-domain> --url localhost:2222 \
  --service-token-id <ID> --service-token-secret <SECRET> &
ssh -i ~/.ssh/deploy -p 2222 orangepi@127.0.0.1 'echo OK'
```
鎺ㄤ竴娆′唬鐮佸埌 main锛岃瀵?GitHub Actions 鐨?Deploy 鏄惁 success锛屾澘瀛愭湇鍔℃槸鍚︽洿鏂般€?
## 7. 鏃ュ父缁存姢

```bash
# 鏌ョ湅鏈嶅姟
systemctl status wol-web
# 鏌ョ湅鏃ュ織
journalctl -u wol-web -f
# 鎵嬪姩閲嶅惎
sudo systemctl restart wol-web
# 鏀归厤缃紙.env锛?sudo nano /opt/wol-web/.env && sudo systemctl restart wol-web
```

## 8. 鍥炴粴

CI 閮ㄧ讲鐨勬槸浜戠缂栬瘧鐨勫浐瀹氱増鏈簩杩涘埗銆傚洖婊氭柟寮忥細
- 鐢?`git revert` 鍥為€€浠ｇ爜鍚?push锛堣Е鍙戦噸鏂伴儴缃叉棫鐗堬級
- 鎴栨墜鍔ㄦ浛鎹?`/opt/wol-web/wol-web` 涓轰笂涓€鐗堜簩杩涘埗骞堕噸鍚?