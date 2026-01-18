# 第17章: PPTP 機能の設定

> 元PDFページ: 346-352

---

第 17 章
PPTP 機能の設定
本機能を使用して  PC と接続するためには、 PC 側には  Microsoft 社 Windows の「仮想プライベートネットワーク」
が必要となります。
17.1 共通の設定
tunnel encapsulation 、tunnel endpoint address 、tunnel endpoint name 、ppp ccp type  コマンドも合わせて参照のこと。
17.1.1 PPTP サーバーを動作させるか否かの設定
[書式 ]
pptp  service  service
no pptp  service  [service ]
[設定値及び初期値 ]
•service
• [設定値 ] :
設定値 説明
on PPTP サーバーとして動作する
off PPTP サーバーとして動作しない
• [初期値 ] : off
[説明 ]
PPTP サーバー機能を動作させるか否かを設定する。
[ノート ]
off に設定すると  PPTP サーバーで使う  TCP のポート番号  1723 を閉じる。デフォルト  off なので、 PPTP サーバーを
起動する場合には、 pptp service  on を設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.1.2 相手先情報番号にバインドされるトンネルインタフェースの設定
[書式 ]
pp bind  interface  [interface  ...]
no pp bind  [interface ]
[設定値及び初期値 ]
•interface
• [設定値 ] :
設定値 説明
tunnelN TUNNEL インターフェース名  (N はインターフェー
ス番号  (tunnel_num ))
tunnelN-tunnelM TUNNEL インターフェースの範囲  (N、M はインター
フェース番号  (tunnel_num ))
• [初期値 ] : -
[説明 ]
選択されている相手先情報番号にバインドされるトンネルインタフェースを指定する。
anonymous インタフェースに対してのみ、複数のトンネルインタフェースが指定できる。
また、連続している複数のトンネルインタフェースの場合は、インタフェース範囲指定が可能である。346 | コマンドリファレンス  | PPTP 機能の設定

[ノート ]
PPTP または  L2TP/IPsec は PP 毎に設定する。
tunnel encapsulation  コマンドで  pptp または  l2tp を設定したトンネルインタフェースをバインドすることによって
PPTP または  L2TP/IPsec で通信することを可能にする。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•tunnel_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.1.3 PPTP の動作タイプの設定
[書式 ]
pptp  service  type type
no pptp  service  type [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
server サーバーとして動作
client クライアントとして動作
• [初期値 ] : server
[説明 ]
PPTP サーバーとして動作するか、 PPTP クライアントとして動作するかを設定する。
[ノート ]
PPTP はサーバー、クライアント方式の接続で、ルーター間で接続する場合には必ず一方がサーバーで、もう一方が
クライアントである必要がある。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.1.4 PPTP ホスト名の設定
[書式 ]
pptp  hostname  name
no pptp  hostname  [name ]
[設定値及び初期値 ]
•name
• [設定値 ] : ホスト名  (64 バイト以下  )
• [初期値 ] :
•なし ( RTX1210 Rev.14.01.26 以降、 RTX830 Rev.15.02.03 以降、および、 Rev.15.04 系以降  )
•機種名 ( 上記以外  )コマンドリファレンス  | PPTP 機能の設定  | 347

[説明 ]
PPTP ホスト名を設定する。
[ノート ]
コマンドで設定したユーザ定義の名前が相手先に通知される。
相手先のルーターには、 show status pp  コマンドの  ' 接続相手先  :' で表示される。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.1.5 PPTP ホスト名の設定
[書式 ]
pptp  vendorname  name
no pptp  vendorname  [name ]
[設定値及び初期値 ]
•name
• [設定値 ] : ベンダー名  (最大  64 文字 /半角、 32 文字 /全角  )
• [初期値 ] : -
[説明 ]
PPTP ベンダー名を設定する。
[ノート ]
本コマンドで設定した値が Start-Control-Connection-Request とStart-Control-Connection-Reply のベンダー名にセット
される。
本コマンドが設定されていないときはベンダー名に空文字がセットされる。
RTX1210 Rev.14.01.26 以降、 RTX830 Rev.15.02.03 以降、および、 Rev.15.04 系以降のファームウェアで使用可能。
それ以外のファームウェアではベンダー名に "YAMAHA Corporation" がセットされる。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.1.6 PPTP パケットのウィンドウサイズの設定
[書式 ]
pptp  window  size size
no pptp  window  size [size]
[設定値及び初期値 ]
•size
• [設定値 ] : パケットサイズ  (1..128)
• [初期値 ] : 32
[説明 ]
受信済みで無応答の  PPTP パケットをバッファに入れることができるパケットの最大数を設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.1.7 PPTP 暗号鍵生成のための要求する認証方式の設定
[書式 ]
pp auth  request  auth [arrive-only]
no pp auth  request  [auth]
[設定値及び初期値 ]
•auth
• [設定値 ] :348 | コマンドリファレンス  | PPTP 機能の設定

設定値 説明
pap PAP
chap CHAP
mschap MSCHAP
mschap-v2 MSCHAP-Version2
chap-pap CHAP と PAP 両方
• [初期値 ] : -
[説明 ]
要求する認証方式を設定します
[ノート ]
PPTP 暗号鍵生成のために認証プロトコルの  MS-CHAP または  MS-CHAPv2 を設定する。通常サーバー側で設定す
る。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.1.8 PPTP 暗号鍵生成のための受け入れ可能な認証方式の設定
[書式 ]
pp auth  accept  auth [auth]
no pp auth  accept  [auth auth]
[設定値及び初期値 ]
•auth
• [設定値 ] :
設定値 説明
pap PAP
chap CHAP
mschap MSCHAP
mschap-v2 MSCHAP-Version2
• [初期値 ] : -
[説明 ]
受け入れ可能な認証方式を設定します。
[ノート ]
PPTP 暗号鍵生成のために認証プロトコルの  MS-CHAP または  MS-CHAPv2 を設定する。通常クライアント側で設
定する。
MacOS 10.2 以降  および  Windows Vista 、Windows 7 をクライアントとして使用する場合は  mschap-v2 を用いる。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.1.9 PPTP のコネクション制御の  syslog を出力するか否かの設定
[書式 ]
pptp  syslog  syslog
no pptp  syslog  [syslog ]
[設定値及び初期値 ]
•syslog
• [設定値 ] :コマンドリファレンス  | PPTP 機能の設定  | 349

設定値 説明
on 出力する
off 出力しない
• [初期値 ] : off
[説明 ]
PPTP のコネクション制御の  syslog を出力するか否かを設定する。
キープアライブ用の  Echo-Request, Echo-Reply については出力されない。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.2 リモートアクセス  VPN 機能
17.2.1 PPTP トンネルの出力切断タイマの設定
[書式 ]
pptp  tunnel  disconnect  time  time
no pptp  tunnel  disconnect  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
1..21474836 秒数
off タイマを設定しない
• [初期値 ] : 60
[説明 ]
選択されている  PPTP トンネルに対して、データパケット無送信の場合、タイムアウトにより  PPTP トンネルを切断
する時間を設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.2.2 PPTP キープアライブの設定
[書式 ]
pptp  keepalive  use use
no pptp  keepalive  use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : on
[説明 ]
トンネルキープアライブを使用するか否かを選択する。
[ノート ]
PPTP トンネルの端点に対して、 PPTP 制御コネクション確認要求  (Echo-Request) を送出して、それに対する  PPTP 制
御コネクション確認要求への応答  (Echo-Reply) で相手先からの応答があるかどうか確認する。応答がない場合に
は、 pptp keepalive interval  コマンドに従った切断処理を行う。350 | コマンドリファレンス  | PPTP 機能の設定

[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.2.3 PPTP キープアライブのログ設定
[書式 ]
pptp  keepalive  log log
no pptp  keepalive  log [log]
[設定値及び初期値 ]
•log
• [設定値 ] :
設定値 説明
on ログにとる
off ログにとらない
• [初期値 ] : off
[説明 ]
トンネルキープアライブをログに取るかどうか選択する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.2.4 PPTP キープアライブを出すインターバルとカウントの設定
[書式 ]
pptp  keepalive  interval  interval  [count ]
no pptp  keepalive  interval  [interval  count ]
[設定値及び初期値 ]
•interval
• [設定値 ] : インターバル  (1..65535)
• [初期値 ] : 30
•count
• [設定値 ] : カウント  (3..100)
• [初期値 ] : 6
[説明 ]
トンネルキープアライブを出すインターバルとダウン検出用のカウントを設定する。
[ノート ]
一度  PPTP 制御コネクション確認要求  (Echo-Request) に対するリプライが返ってこないのを検出したら、その後の監
視タイマは  1 秒に短縮される。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
17.2.5 PPTP 接続において暗号化の有無により接続を許可するか否かの設定
[書式 ]
ppp ccp no-encryption  mode
no ppp ccp no-encryption  [mode ]
[設定値及び初期値 ]
•mode
• [設定値 ] :
設定値 説明
reject 暗号化なしでは接続拒否
accept 暗号化なしでも接続許可コマンドリファレンス  | PPTP 機能の設定  | 351

• [初期値 ] : accept
[説明 ]
MPPE(Microsoft Point-to-Point Encryption) の暗号化がネゴシエーションされないときの動作を設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830352 | コマンドリファレンス  | PPTP 機能の設定

