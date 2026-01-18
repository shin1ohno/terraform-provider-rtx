# 第22章: RADIUS の設定

> 元PDFページ: 397-400

---

第 22 章
RADIUS の設定
ISDN 接続のための認証とアカウントを  RADIUS サーバーを利用して管理できます。 PPTP 接続のための認証とアカ
ウントの管理はサポートされません。
22.1 RADIUS による認証を使用するか否かの設定
[書式 ]
radius  auth  auth
no radius  auth  [auth]
[設定値及び初期値 ]
•auth
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : off
[説明 ]
anonymous に対して何らかの認証を要求する設定の場合に、相手から受け取ったユーザネーム  (PAP であれば
UserID、CHAP であれば  NAME) が、自分で持つユーザネーム  (pp auth username  コマンドで指定  ) の中に含まれて
いない場合には  RADIUS サーバーに問い合わせるか否かを設定する。
[ノート ]
RADIUS による認証と  RADIUS によるアカウントは独立して使用できる。
サポートしているアトリビュートについては、 WWW サイトのドキュメント <https://www.rtpro.yamaha.co.jp> を参照
すること。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
22.2 RADIUS によるアカウントを使用するか否かの設定
[書式 ]
radius  account  account
no radius  account  [account ]
[設定値及び初期値 ]
•account
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : off
[説明 ]
RADIUS によるアカウントを使用するか否かを設定する。
[ノート ]
RADIUS による認証と  RADIUS によるアカウントは独立して使用できる。コマンドリファレンス  | RADIUS の設定  | 397

サポートしているアトリビュートについては、 WWW サイトのドキュメント <https://www.rtpro.yamaha.co.jp> を参照
すること。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
22.3 RADIUS サーバーの指定
[書式 ]
radius  server  ip1 [ip2]
no radius  server  [ip1 [ip2]]
[設定値及び初期値 ]
•ip1
• [設定値 ] : RADIUS サーバー ( 正 ) の IP アドレス  (IPv6 アドレス可  )
• [初期値 ] : -
•ip2
• [設定値 ] : RADIUS サーバー ( 副 ) の IP アドレス  (IPv6 アドレス可  )
• [初期値 ] : -
[説明 ]
RADIUS サーバーを設定する。 2 つまで指定でき、最初のサーバーから返事をもらえない場合は、 2 番目のサーバー
に問い合わせを行う。
[ノート ]
RADIUS には認証とアカウントの  2 つの機能があり、それぞれのサーバーは radius auth server /radius account server
コマンドで個別に設定できる。 radius server  コマンドでの設定は、これら個別の設定が行われていない場合に有効
となり、認証、アカウントいずれでも用いられる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
22.4 RADIUS 認証サーバーの指定
[書式 ]
radius  auth  server  ip1 [ip2]
no radius  auth  server  [ip1 [ip2]]
[設定値及び初期値 ]
•ip1
• [設定値 ] : RADIUS 認証サーバー ( 正 ) の IP アドレス  (IPv6 アドレス可  )
• [初期値 ] : -
•ip2
• [設定値 ] : RADIUS 認証サーバー ( 副 ) の IP アドレス  (IPv6 アドレス可  )
• [初期値 ] : -
[説明 ]
RADIUS 認証サーバーを設定する。 2 つまで指定でき、最初のサーバーから返事をもらえない場合は、 2 番目のサー
バーに問い合わせを行う。
[ノート ]
このコマンドで  RADIUS 認証サーバーの  IP アドレスが指定されていない場合は、 radius server  コマンドで指定した
IP アドレスを認証サーバーとして用いる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
22.5 RADIUS アカウントサーバーの指定
[書式 ]
radius  account  server  ip1 [ip2]
no radius  account  server  [ip1 [ip2]]398 | コマンドリファレンス  | RADIUS の設定

[設定値及び初期値 ]
•ip1
• [設定値 ] : RADIUS アカウントサーバー ( 正 ) の IP アドレス  (IPv6 アドレス可  )
• [初期値 ] : -
•ip2
• [設定値 ] : RADIUS アカウントサーバー ( 副 ) の IP アドレス  (IPv6 アドレス可  )
• [初期値 ] : -
[説明 ]
RADIUS アカウントサーバーを設定する。 2 つまで指定でき、最初のサーバーから返事をもらえない場合は、 2 番目
のサーバーに問い合わせを行う。
[ノート ]
このコマンドで  RADIUS アカウントサーバーの  IP アドレスが指定されていない場合は、 radius server  コマンドで指
定した  IP アドレスをアカウントサーバーとして用いる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
22.6 RADIUS 認証サーバーの  UDP ポートの設定
[書式 ]
radius  auth  port port_num
no radius  auth  port [port_num ]
[設定値及び初期値 ]
•port_num
• [設定値 ] : UDP ポート番号
• [初期値 ] : 1645
[説明 ]
RADIUS 認証サーバーの  UDP ポート番号を設定する
[ノート ]
RFC2138 ではポート番号として  1812 を使うことになっている。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
22.7 RADIUS アカウントサーバーの  UDP ポートの設定
[書式 ]
radius  account  port port_num
no radius  account  port [port_num ]
[設定値及び初期値 ]
•port_num
• [設定値 ] : UDP ポート番号
• [初期値 ] : 1646
[説明 ]
RADIUS アカウントサーバーの  UDP ポート番号を設定する。
[ノート ]
RFC2138 ではポート番号として  1813 を使うことになっている。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | RADIUS の設定  | 399

22.8 RADIUS シークレットの設定
[書式 ]
radius  secret  secret
no radius  secret  [secret ]
[設定値及び初期値 ]
•secret
• [設定値 ] : シークレット文字列  (16文字以内 )
• [初期値 ] : -
[説明 ]
RADIUS シークレットを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
22.9 RADIUS 再送信パラメータの設定
[書式 ]
radius  retry  count  time
no radius  retry  [count  time]
[設定値及び初期値 ]
•count
• [設定値 ] : 再送回数  (1..10)
• [初期値 ] : 4
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
RADIUS パケットの再送回数とその時間間隔を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830400 | コマンドリファレンス  | RADIUS の設定

