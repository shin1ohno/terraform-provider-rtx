# 第20章: SIP 機能の設定

> 元PDFページ: 360-377

---

第 20 章
SIP 機能の設定
20.1 共通の設定
20.1.1 SIP を使用するか否かの設定
[書式 ]
sip use use
no sip use
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
off 使用しない
on 使用する
• [初期値 ] : off
[説明 ]
SIP プロトコルを使用するか否かを設定する。
[ノート ]
on から  off への設定の変更は再起動後有効となる。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.2 SIP の session-timer 機能のタイマ値の設定
[書式 ]
sip session  timer  time [update= update ] [refresher= refresher ]
no sip session  timer
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
秒数  (60..540)
0 session-timer 機能を利用しない
• [初期値 ] : 0
•update
• [設定値 ] :
設定値 説明
on UPDATE メソッドを使用する
off UPDATE メソッドを使用しない
• [初期値 ] : off
•refresher
• [設定値 ] :360 | コマンドリファレンス  | SIP 機能の設定

設定値 説明
none refresher パラメータを設定しない
uac refresher パラメータに  uac を設定する
uas refresher パラメータに  uas を設定する
• [初期値 ] : uac
[説明 ]
SIP の session-timer 機能のタイマ値を設定する。
SIP の通話中に相手が停電などにより突然落ちた場合にタイマにより自動的に通話を切断する。
update  を on に設定すれば、発信時に  session-timer 機能において  UPDATE メソッドを使用可能とする。
refresher  を none に設定した時は  refresher パラメータを設定せず、 uac/uas を設定した時はそれぞれのパラメータ値で
発信する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.3 SIP による発信時に使用する IP プロトコルの選択
[書式 ]
sip ip protocol  protocol
no sip ip protocol
[設定値及び初期値 ]
•protocol
• [設定値 ] :
設定値 説明
udp UDP を使用
tcp TCP を使用
• [初期値 ] : udp
[説明 ]
SIP による発信時の呼制御に使用する  IP プロトコルを選択する。
[ノート ]
着信した場合は、この設定に関わらず、受信したプロトコルで送信を行なう。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.4 SIP による発信時に  100rel をサポートするか否かの設定
[書式 ]
sip 100rel  switch
no sip 100rel
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 100rel をサポートする
off 100rel をサポートしない
• [初期値 ] : offコマンドリファレンス  | SIP 機能の設定  | 361

[説明 ]
SIP の発信時に  100rel(RFC3262) をサポートするか否かを設定する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.5 送信する  SIP パケットに  User-Agent ヘッダを付加する設定
[書式 ]
sip user agent  sw [user-agent ]
no sip user agent
[設定値及び初期値 ]
•sw
• [設定値 ] :
設定値 説明
on 付加する
off 付加しない
• [初期値 ] : off
•user-agent
• [設定値 ] : ヘッダに記述する文字列
• [初期値 ] : -
[説明 ]
送信する  SIP パケットに  User-Agent ヘッダを付加することができる。
付加する文字列は、 user-agent  パラメータにて設定することが可能であるが、 64 文字以内で  ASCII 文字のみ設定可
能である。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.6 SIP による着信時の  INVITE に refresher 指定がない場合の設定
[書式 ]
sip arrive  session  timer  refresher  refresher
no sip arrive  session  timer  refresher
[設定値及び初期値 ]
•refresher
• [設定値 ] :
設定値 説明
uac refresher=uac と指定する
uas refresher=uas と指定する
• [初期値 ] : uac
[説明 ]
SIP による着信時の  INVITE が refresher を指定していない場合に  UAC/UAS を指定できる。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.7 SIP による着信時に  P-N-UAType ヘッダをサポートするか否かの設定
[書式 ]
sip arrive  ringing  p-n-uatype  switch
no sip arrive  ringing  p-n-uatype362 | コマンドリファレンス  | SIP 機能の設定

[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on P-N-UAType ヘッダを付加する
off P-N-UAType ヘッダを付加しない
• [初期値 ] : off
[説明 ]
SIP による着信時に送信する  Ringing レスポンスに、 P-N-UAType ヘッダを付加するか否かを設定する。
[ノート ]
設定はすべての着信に適用される。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.8 SIP による着信時のセッションタイマーのリクエストを設定
[書式 ]
sip arrive  session  timer  method  method
no sip arrive  session  timer  method  [method ]
[設定値及び初期値 ]
•method
• [設定値 ] :
設定値 説明
auto 自動的に判断する
invite INVITE のみを使用する
• [初期値 ] : auto
[説明 ]
SIP による着信時にセッションタイマー機能で使用するリクエストを設定する。
auto に設定した場合には  UPDATE, INVITE ともに使用でき、発信側またはサーバで  UPDATE に対応していれば
UPDATE を使用する。
invite に設定した場合には、発信側またはサーバで  UPDATE に対応していてもこれを使用せずに動作する。
UPDATE のみを使用する設定はできない。
また、サーバ毎に設定することできないため、全ての着信でこの設定が有効となる。
発信の場合は、 sip server session timer  または  sip session timer  の update  オプションで設定できる。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.9 SIP 着信時にユーザー名を検証するか否かの設定
[書式 ]
sip arrive  address  check  switch
no sip arrive  address  check
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on ユーザ名を検証するコマンドリファレンス  | SIP 機能の設定  | 363

設定値 説明
off ユーザ名を検証しない
• [初期値 ] : on
[説明 ]
SIPサーバーの設定をした場合に、着信時の Request-URI が送信した REGISTER のContactヘッダの内容と一致する
かを検証するか否かを設定する。
SIPを利用した V oIP機能において、 SIPサーバーを利用する設定と Peer to Peer で利用する設定を併用する場合は off
にする。
また、 SIPサーバーに RTV01を利用する場合にも offにする。
[ノート ]
この検証は  sip server  設定がある場合に有効となる。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.10 着信可能なポートがない場合に返す  SIP のレスポンスコードの設定
[書式 ]
sip response  code  busy  code
no sip response  code  busy
[設定値及び初期値 ]
•code  : レスポンスコード
• [設定値 ] :
設定値 説明
486 486 を返す
503 503 を返す
• [初期値 ] : 486
[説明 ]
SIP 着信時に、ビジーで着信できない場合に返すレスポンスコードを設定する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.11 SIP で使用する IPアドレスの設定
[書式 ]
sip outer  address  ipaddress
no sip outer  address
[設定値及び初期値 ]
•ipaddress
• [設定値 ] :
設定値 説明
auto 自動設定
IP アドレス IP アドレス
• [初期値 ] : auto
[説明 ]
SIP で使用する IPアドレスを設定する。  RTP/RTCP もこの値が使用される。
[ノート ]
初期設定のまま使用する事を推奨する。364 | コマンドリファレンス  | SIP 機能の設定

[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.1.12 SIP メッセージのログを記録するか否かの設定
[書式 ]
sip log switch
no sip log
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on SIP メッセージのログを記録する
off SIP メッセージのログを記録しない
• [初期値 ] : off
[説明 ]
SIP メッセージのログを  DEBUG レベルのログに記録するか否かを設定する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.2 SIP サーバー毎の設定
20.2.1 SIP サーバーの設定
[書式 ]
sip server  number  address  type protocol  sip_uri  [username  [password ]]
no sip server  number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•address
• [設定値 ] : SIP サーバーの  IP アドレス
• [初期値 ] : -
•type
• [設定値 ] :
• register
• no-register
• [初期値 ] : -
•protocol
• [設定値 ] :
設定値 説明
tcp TCP プロトコル
udp UDP プロトコル
• [初期値 ] : -
•sip_url
• [設定値 ] : SIP アドレス
• [初期値 ] : -
•username
• [設定値 ] : ユーザ名
• [初期値 ] : -コマンドリファレンス  | SIP 機能の設定  | 365

•password
• [設定値 ] : パスワード
• [初期値 ] : -
[説明 ]
SIP サーバー設定を追加または削除する。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.2 SIP サーバー毎の  session-timer 機能のタイマ値の設定
[書式 ]
sip server  session  timer  number  time [update= update ] [refresher= refresher ]
no sip server  session  timer  number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•time
• [設定値 ] :
•秒数 (60..540)
• 0 ... session-timer 機能を利用しない
• [初期値 ] : -
•update
• [設定値 ] :
設定値 説明
on UPDATE メソッドを使用する
off UPDATE メソッドを使用しない
• [初期値 ] : -
•refresher
• [設定値 ] :
設定値 説明
none refresher パラメータを設定しない
uac refresher パラメータに  uac を設定する
uas refresher パラメータに  uas を設定する
• [初期値 ] : -
[説明 ]
SIP サーバー毎の  session-timer 機能のタイマ値を設定する。
SIP の通話中に相手が停電などにより突然落ちた場合にタイマにより自動的に通話を切断する。
サーバーが  session-timer に対応していれば、 端末が 2 台同時に突然落ちてもサーバーでの呼の持ち切りを防ぐことが
できる。
update  を on に設定すれば、発信時に  session-timer 機能において  UPDATE メソッドを使用可能とする。
refresher  を none に設定した時は  refresher パラメータを設定せず、 uac/uas を設定した時はそれぞれのパラメータ値で
発信する。
[適用モデル ]
RTX5000, RTX3510, RTX3500366 | コマンドリファレンス  | SIP 機能の設定

20.2.3 SIP サーバー毎の代表 SIP アドレスの設定
[書式 ]
sip server  pilot  address  number  sipaddress
no sip server  pilot  address  number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•sipaddress
• [設定値 ] : 代表  SIP アドレス
• [初期値 ] : -
[説明 ]
SIP サーバー経由の発信時に、 INVITE リクエストの  P-Preferred-Identity ヘッダに設定した代表  SIP アドレスを入れて
発信する。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.4 SIP サーバー毎の先頭に付加された  184/186 の扱いの設定
[書式 ]
sip server  privacy  number  switch  [pattern ]
no sip server  privacy  number  switch  [pattern ]
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
off ダイヤルされたそのままの番号で発信する
always-off ダイヤルされた番号から  184 / 186 を取り除き、常に
「通知」で発信する
always-on ダイヤルされた番号から  184 / 186 を取り除き、常に
「非通知」で発信する
default-off ダイヤルされた番号から  184 / 186 を取り除き、 184 が
付加されている場合には「非通知」で、それ以外の場
合には「通知」で発信する。
default-on ダイヤルされた番号から  184 / 186 を取り除き、 186 が
付加されている場合には「通知」で、それ以外の場合
には「非通知」で発信する。
• [初期値 ] : off
•pattern
• [設定値 ] :
設定値 説明
sip-privacy draft-ietf-sip-privacy-01 に従って発信者番号の通知  /
非通知を行なう。
rfc3325 RFC3325 に従って発信者番号の通知  / 非通知を行な
う。
as-is ダイヤルされた番号に  184 / 186 を付加して発信す
る。コマンドリファレンス  | SIP 機能の設定  | 367

• [初期値 ] : -
[説明 ]
ダイヤルされた番号の先頭に付加された  184 / 186 をどのように取り扱うかを指定する。
各 pattern  パラメータで指定した方式に従って、ダイヤルされた番号を処理する。 pattern  パラメータを省略した場合
は、 draft-ietf-sip-privacy-01 に従って、ダイヤルされた番号を処理する。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.5 SIP サーバー毎の発信時に使用する自己 SIP ディスプレイ名の設定
[書式 ]
sip server  display  name  number  displayname
no sip server  display  name  number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•displayname
• [設定値 ] : ディスプレイ名
• [初期値 ] : -
[説明 ]
SIP サーバー毎の発信時に使用される自己  SIP ディスプレイ名を設定する。
[ノート ]
空白を含むディスプレイ名を設定する場合、 "" で囲む必要がある。
漢字を設定する場合は、シフト JIS コードで設定を行なう。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.6 SIP サーバー毎の発信時の相手  SIP アドレスのドメイン名の設定
[書式 ]
sip server  call remote  domain  number  domain
no sip server  call remote  domain  number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•domain
• [設定値 ] : ドメイン名
• [初期値 ] : -
[説明 ]
SIP サーバー経由の発信時に、相手の  SIP アドレスの  host 部分を設定したドメイン名にして発信する。
ドメイン名の長さは  58 文字まで設定できる。
なお、 ドメイン名として使用可能な文字は、 アルファベット、 数字、 ハイフン、 ピリオド、 コロン、 カッコ [ ] のみである。
ドメイン名を設定しない場合には、 sip server  コマンドの  SIP-URI の host 部分と同じドメイン名にして発信する。
[適用モデル ]
RTX5000, RTX3510, RTX3500368 | コマンドリファレンス  | SIP 機能の設定

20.2.7 SIP サーバー毎の発信時に  100rel をサポートするか否かの設定
[書式 ]
sip server  100rel  number  switch
no sip server  100rel  number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 100rel をサポートする
off 100rel をサポートしない
• [初期値 ] : off
[説明 ]
SIP サーバー経由の発信時に  100rel(RFC3262) をサポートするか否かを設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.8 SIP サーバー毎の  REGISTER リクエストの更新間隔の設定
[書式 ]
sip server  register  timer  server= number  OK_time  NG_time
no sip server  register  timer  server= number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•OK_time
• [設定値 ] : 通常時更新間隔  (10..120 ( 分))
• [初期値 ] : 30
•NG_time
• [設定値 ] : 異常時更新間隔  (1..60 (分))
• [初期値 ] : 5
[説明 ]
SIP サーバーに  REGISTER リクエストを送信する間隔を設定する。
正常に更新されている場合には通常時更新間隔毎に更新する。サーバーからエラーが返されたり、サーバーから応
答が無い場合には、異常時更新間隔毎に更新する。また、この時の  Expires ヘッダは通常時更新間隔を  2 倍して秒に
直した値で送信する。しかし、サーバーから  Expires の指定があった場合はその値に従って、指定された値の半分の
時間で通常時の更新を行なう。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.9 SIP サーバー毎の  REGISTER リクエストの  Request-URI の設定
[書式 ]
sip server  register  request-uri  number  sip_address
no sip server  register  request-uri  number
[設定値及び初期値 ]
•numberコマンドリファレンス  | SIP 機能の設定  | 369

• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•sip_address
• [設定値 ] : Request-URI
• [初期値 ] : -
[説明 ]
SIP サーバーに送信する  REGISTER リクエストの  Request-URI を設定する。
設定しない場合は、 sip server  コマンドで設定した  SIP-URI の host 部分を入れて  REGISTER リクエストを送信する。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.10 SIP サーバー毎の  REGISTER リクエストの  Contact ヘッダに付加する  q 値の設定
[書式 ]
sip server  qvalue  number  value
no sip server  qvalue  number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•value
• [設定値 ] :
設定値 説明
q 値 (0.001..1.000)
0 q 値を付加しない
• [初期値 ] : 0
[説明 ]
SIP サーバーへ接続する時に送信する  REGISTER リクエストの  Contact ヘッダに付加する  q 値を設定する。 0.001 単
位で設定可能。
同じアカウントで同時に複数の端末から接続が許されている  SIP サーバーを利用する時に、この設定により着信す
る優先順位を  SIP サーバーに通知することが可能となる。数値が大きい方が優先される。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.11 自分自身の  SIP アドレスへの発信を許可するかどうかの設定
[書式 ]
sip server  call own permit  server= number  sw
no sip server  call own permit  server= number
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•sw
• [設定値 ] :
設定値 説明
on 許可する
off 許可しない370 | コマンドリファレンス  | SIP 機能の設定

• [初期値 ] : off
[説明 ]
To, From が同じ  SIP アドレスとなるような発信を許可するか否かを設定する。
この機能を利用して正常に発信ができるのは、 Call-ID や tag 等の乱数値を発信側と着信側で別の値を付加して管理
する  SIP サーバーを利用する場合だけである。
そのため、通常は  off で運用する。
[適用モデル ]
RTX5000, RTX3510, RTX3500
20.2.12 SIP サーバー毎の  REGISTER リクエストの  Contact ヘッダーの設定
[書式 ]
sip server  register  contact  mode  number  mode
no sip server  register  contact  mode  number  [mode ]
[設定値及び初期値 ]
•number
• [設定値 ] : SIP サーバーの登録番号  (1..65535)
• [初期値 ] : -
•mode
• [設定値 ] : 動作モード  (1..3)
• [初期値 ] : 1
[説明 ]
SIP サーバーに送信する  REGISTER リクエストの  Contact ヘッダーに関する動作を設定する。
[適用モデル ]
RTX3510
20.3 NGN 機能の設定
データコネクトを利用して拠点間接続を行うにはトンネルインタフェースを利用します。トンネリングの章や
IPsecの設定の章を参照してください。
20.3.1 NGN 網に接続するインタフェースの設定
[書式 ]
ngn type interface  type
no ngn type interface  [type]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース
• [初期値 ] : -
•type
• [設定値 ] :
設定値 説明
off NGN 網のサービスを使用しない
ntt NTT 東日本または NTT 西日本が提供する NGN 網を
使用する
• [初期値 ] : off
[説明 ]
NGN 網に接続するインタフェースを設定する。コマンドリファレンス  | SIP 機能の設定  | 371

[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.2 NGN 網を介したトンネルインタフェースの切断タイマの設定
[書式 ]
tunnel  ngn disconnect  time  time
no tunnel  ngn disconnect  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
1..21474836 秒数
off タイマを設定しない
• [初期値 ] : 60
[説明 ]
NGN網を介したトンネルインタフェースのデータ送受信がない場合の切断までの時間を設定する。  offに設定した
場合は切断しない。
[ノート ]
通信中の変更は無効。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.3 NGN 網を介したトンネルインタフェースの帯域幅の設定
[書式 ]
tunnel  ngn bandwidth  bandwidth  [arrivepermit= switch ]
no tunnel  ngn bandwidth  [bandwidth  arrivepermit= switch ]
[設定値及び初期値 ]
•bandwidth
• [設定値 ] :
設定値 説明
64k 64kbps
512k 512kbps
1m 1Mbps
1k..1000m 帯域
• [初期値 ] : 1m
•switch
• [設定値 ] :
設定値 説明
on 帯域の設定と一致しない着信も許可する
off 帯域の設定と一致した着信のみ許可する
• [初期値 ] : on
[説明 ]
NGN網を介したトンネルインタフェースの帯域幅を設定した値にする。
帯域の設定が一致しない着信について、 arrivepermit オプションが offの場合は着信せず、 onの場合は着信する。372 | コマンドリファレンス  | SIP 機能の設定

[ノート ]
通信中の変更は無効。
bandwidth への任意の帯域の指定は、 RTX5000 / RTX3500 Rev.14.00.20 以前、 RTX1210 Rev.14.01.15 以前では不可能。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.4 NGN 網を介したトンネルインタフェースの着信許可の設定
[書式 ]
tunnel  ngn arrive  permit  permit
no tunnel  ngn arrive  permit  [permit ]
[設定値及び初期値 ]
•permit
• [設定値 ] :
設定値 説明
on 許可する
off 許可しない
• [初期値 ] : on
[説明 ]
選択されている相手からの着信を許可するか否かを設定する。
[ノート ]
tunnel ngn arrive permit 、tunnel ngn call permit  コマンドとも  off を設定した場合は通信できない。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.5 NGN 網を介したトンネルインタフェースの発信許可の設定
[書式 ]
tunnel  ngn call permit  permit
no tunnel  ngn call permit  [permit ]
[設定値及び初期値 ]
•permit
• [設定値 ] :
設定値 説明
on 許可する
off 許可しない
• [初期値 ] : on
[説明 ]
選択されている相手への発信を許可するか否かを設定する。
[ノート ]
tunnel ngn arrive permit 、tunnel ngn call permit  コマンドとも  off を設定した場合は通信できない。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | SIP 機能の設定  | 373

20.3.6 NGN 網を介したトンネルインタフェースで使用する LANインタフェースの設定
[書式 ]
tunnel  ngn interface  lan
no tunnel  ngn interface  [lan]
[設定値及び初期値 ]
•lan
• [設定値 ] :
設定値 説明
auto 自動設定
LAN インタフェース名 LAN ポート
• [初期値 ] : auto
[説明 ]
NGN網を介したトンネルインタフェースで使用する LANインタフェースを設定する。
autoに設定した時はトンネルインタフェースで設定した電話番号を利用して、使用する LANインタフェースを決定
する。
追加番号を使用する場合や HGW配下で使用する場合に設定する。
[ノート ]
通信中の変更は無効。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.7 NGN 網を介したトンネルインタフェースで接続に失敗した場合に接続を試みる相手番号の設
定
[書式 ]
tunnel  ngn fallback  remote_tel  ...
no tunnel  ngn fallback  [remote_tel  ...]
[設定値及び初期値 ]
•remote_tel
• [設定値 ] : 相手電話番号
• [初期値 ] : -
[説明 ]
NGN網を介したトンネルインタフェースで使用する相手番号は、 ipsec ike remote name  コマンドや  tunnel endpoint
name  コマンドで設定した番号に対して発信するが、これが何らかの原因で接続できなかった場合に、設定された番
号に対して発信する。
設定は最大 7個まで可能で、接続に失敗すると設定された順番に次の番号を用いて接続を試みる。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.8 NGN 電話番号を RADIUS で認証するか否かの設定
[書式 ]
tunnel  ngn radius  auth  use
no tunnel  ngn radius  auth
[設定値及び初期値 ]
•use
• [設定値 ] :374 | コマンドリファレンス  | SIP 機能の設定

設定値 説明
on 認証する
off 認証しない
• [初期値 ] : off
[説明 ]
データコネクトを利用した拠点間接続において、着信を受けたときに発信元の NGN電話番号を RADIUS で認証す
るか否かを設定する。
[ノート ]
トンネルインタフェースが選択されている時にのみ使用できる。
トンネルに相手の電話番号が設定されている場合は RADIUS 認証を行わない。
以下のコマンドが正しく設定されている必要がある。
•radius account
•radius account server
•radius account port
•radius secret
•ngn radius auth password
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.9 NGN 電話番号を RADIUS で認証するときに使用するパスワードの設定
[書式 ]
ngn radius  auth  password  password
no ngn radius  auth  password
[設定値及び初期値 ]
•password
• [設定値 ] : パスワード
• [初期値 ] : -
[説明 ]
NGN電話番号を RADIUS で認証するときに使用するパスワードを設定する。 NGN電話番号をユーザー名、当コマ
ンドで設定した文字列をパスワードとして RADIUS サーバーに問い合わせを行う。
PASSWORD に使用できる文字は半角英数字および記号  (7bit ASCII Code で表示可能なもの ) で、 文字列の長さは 0文
字以上 64文字以下となる。
[ノート ]
当コマンドが設定されていない場合は、 NGN電話番号を RADIUS で認証することができない。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.10 NGN 網への発信時に RADIUS アカウンティングを使用するか否かの設定
[書式 ]
ngn radius  account  caller  use
no ngn radius  account  caller
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : offコマンドリファレンス  | SIP 機能の設定  | 375

[説明 ]
NGN網への発信時に RADIUS アカウンティングを使用するか否かを設定する。
[ノート ]
RADIUS アカウンティングサーバーに関する以下のコマンドが正しく設定されている必要がある。
•radius account
•radius account server
•radius account port
•radius secret
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.11 NGN 網からの着信時に RADIUS アカウンティングを使用するか否かの設定
[書式 ]
ngn radius  account  callee  use
no ngn radius  account  callee
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : off
[説明 ]
NGN網からの着信時に RADIUS アカウンティングを使用するか否かを設定する。
[ノート ]
RADIUS アカウンティングサーバーに関する以下のコマンドが正しく設定されている必要がある。
•radius account
•radius account server
•radius account port
•radius secret
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.12 NGN 網を介したリナンバリング発生時に LANインターフェースを一時的にリンクダウンす
るか否かの設定
[書式 ]
ngn renumbering  link-refresh  switch
no ngn renumbering  link-refresh  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on リナンバリング発生時、 LANインターフェースを一
時的にリンクダウンする
off リナンバリング発生時、 取得したプレフィックスに変
更がない場合は、 LANインターフェースをリンクダ
ウンしない
• [初期値 ] : on376 | コマンドリファレンス  | SIP 機能の設定

[説明 ]
NGN網を介したリナンバリングが発生した時、 LANインターフェースを一時的にリンクダウンするか否かを設定す
る。
LANインターフェースを一時的にリンクダウンさせることにより、 DHCPv6-PD/RA プロキシの配下のより多くの端
末に対して、 IPv4/IPv6 アドレスの再取得を促し、リナンバリング後も通信を継続できるようにする。
このコマンドを onに設定した場合は、 NGN網を介したリナンバリングの発生時、取得したプレフィックスに変更
がないときでも LANインターフェースを一時的にリンクダウンする。 offに設定した場合は、取得したプレフィッ
クスに変更がないときはリンクダウンしない。
[ノート ]
RTX1210 は Rev.14.01.16 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.21 以降で使用可能。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
20.3.13 NGN 網接続情報の表示
[書式 ]
show  status  ngn
[説明 ]
NGN 網への接続状態を表示する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | SIP 機能の設定  | 377

