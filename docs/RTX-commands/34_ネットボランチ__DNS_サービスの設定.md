# 第34章: ネットボランチ  DNS サービスの設定

> 元PDFページ: 537-545

---

第 34 章
ネットボランチ  DNS サービスの設定
ネットボランチ  DNS とは、一種のダイナミック  DNS 機能であり、ルーターの  IP アドレスをヤマハが運営するネッ
トボランチ  DNS サーバーに希望の名前で登録することができます。そのため、動的  IP アドレス環境でのサーバー
公開や拠点管理などに用いることができます。 IP アドレスの登録、更新などの手順には独自のプロトコルを用いる
ため、他のダイナミック  DNS サービスとの互換性はありません。
ヤマハが運営するネットボランチ  DNS サーバーは現時点では無料、無保証の条件で運営されています。利用料金は
必要ありませんが、ネットボランチ  DNS サーバーに対して名前が登録できること、および登録した名前が引けるこ
とは保証できません。また、ネットボランチ  DNS サーバーは予告無く停止することがあることに注意してくださ
い。
ネットボランチ  DNS には、ホストアドレスサービスと電話番号サービスの  2 種類がありますが、本書で記述するモ
デルでは電話番号サービスは利用できません。
ネットボランチ  DNS では、個々の  RT シリーズ、ネットボランチシリーズルーターを  MAC アドレスで識別してい
るため、機器の入れ換えなどをした場合には同じ名前がそのまま利用できる保証はありません。
34.1 ネットボランチ  DNS サービスの使用の可否
[書式 ]
netvolante-dns  use interface  switch
netvolante-dns  use pp switch
no netvolante-dns  use interface  [switch ]
no netvolante-dns  use pp [switch ]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
auto 自動更新する
off 自動更新しない
• [初期値 ] : auto
[説明 ]
ネットボランチ  DNS サービスを使用するか否かを設定する。
IP アドレスが更新された時にネットボランチ  DNS サーバーに自動で  IP アドレスを更新する。
[ノート ]
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.2 ネットボランチ  DNS サーバーへの手動更新
[書式 ]
netvolante-dns  go interface
netvolante-dns  go pp peer_num
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名コマンドリファレンス  | ネットボランチ  DNS サービスの設定  | 537

• [初期値 ] : -
•peer_num
• [設定値 ] : 相手先情報番号
• [初期値 ] : -
[説明 ]
ネットボランチ  DNS サーバーに手動で  IP アドレスを更新する。
[ノート ]
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.3 ネットボランチ  DNS サーバーからの削除
[書式 ]
netvolante-dns  delete  go interface  [host]
netvolante-dns  delete  go pp peer_num  [host]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名
• [初期値 ] : -
•peer_num
• [設定値 ] : 相手先情報番号
• [初期値 ] : -
•host
• [設定値 ] : ホスト名
• [初期値 ] : -
[説明 ]
登録した  IP アドレスをネットボランチ  DNS サーバーから削除する。
インタフェースの後にホスト名を指定することで、指定したホスト名のみを削除可能。
[ノート ]
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -538 | コマンドリファレンス  | ネットボランチ  DNS サービスの設定

ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 150
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.4 ネットボランチ  DNS サービスで使用するポート番号の設定
[書式 ]
netvolante-dns  port port
no netvolante-dns  port [port]
[設定値及び初期値 ]
•port
• [設定値 ] : ポート番号  (1..65535)
• [初期値 ] : 2002
[説明 ]
ネットボランチ  DNS サービスで使用するポート番号を設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.5 ネットボランチ  DNS サーバーに登録済みのホスト名一覧を取得
[書式 ]
netvolante-dns  get hostname  list interface
netvolante-dns  get hostname  list pp peer_num
netvolante-dns  get hostname  list all
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名
• [初期値 ] : -
•peer_num
• [設定値 ] : 相手先情報番号
• [初期値 ] : -
• all : すべてのインタフェース
• [初期値 ] : -
[説明 ]
ネットボランチ  DNS サーバーに登録済みのホスト名一覧を取得し、表示する。
[ノート ]
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150コマンドリファレンス  | ネットボランチ  DNS サービスの設定  | 539

[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.6 ホスト名の登録
[書式 ]
netvolante-dns  hostname  host interface  [server= server_num ] host [duplicate]
netvolante-dns  hostname  host pp [server= server_num ] host [duplicate]
netvolante-dns  hostname  host interface  [server= server_num ] host ipv6 address  [ipv6_address ] [duplicate]
netvolante-dns  hostname  host pp [server= server_num ] host ipv6 address  [ipv6_address ] [duplicate]
netvolante-dns  hostname  host interface  [server= server_num ] host ipv6 prefix  [ipv6_prefix ] [duplicate]
netvolante-dns  hostname  host pp [server= server_num ] host ipv6 prefix  [ipv6_prefix ] [duplicate]
no netvolante-dns  hostname  host interface  [server= server_num ] [host [duplicate]]
no netvolante-dns  hostname  host pp [server= server_num ] [host [duplicate]]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名
• [初期値 ] : -
•server_num
• [設定値 ] :
設定値 説明
1または 2 サーバ番号
省略 省略時は 1が指定されたものとみなす
• [初期値 ] : -
•host
• [設定値 ] : ホスト名  (63 文字以内  )
• [初期値 ] : -
•ipv6_address
• [設定値 ] : IPv6アドレス
• [初期値 ] : -
•ipv6_prefix
• [設定値 ] : IPv6プレフィックス
• [初期値 ] : -
[説明 ]
ネットボランチ  DNS サービス  ( ホストアドレスサービス  ) で使用するホスト名を設定する。ネットボランチ  DNS
サーバーから取得されるホスト名は、 『 ( ホスト名  ).( サブドメイン  ).netvolante.jp 』という形になる。 ( ホスト名  ) は
このコマンドで設定した名前となり、 ( サブドメイン  ) はネットボランチ  DNS サーバーから割り当てられる。 ( サブ
ドメイン  ) をユーザが指定することはできない。
このコマンドを一番最初に設定する際は、 ( ホスト名  ) 部分のみを設定する。ネットボランチ  DNS サーバーに対し
ての登録・更新が成功すると、コマンドが上記の完全な  FQDN の形になって保存される。
duplicate を付加すると、 1 台のルーターで異なるインタフェースに同じ名前を登録できる。
host に使用できる文字は半角英数字およびハイフン (-)で、  63 文字以内で指定する。
[ノート ]
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
ipv6 address および ipv6 prefix は RTX5000 / RTX3500 Rev.14.00.26 以降、 RTX1210 Rev.14.01.33 以降、 RTX830
Rev.15.02.09 以降、および、 Rev.15.04 系以降のファームウェアで使用可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830540 | コマンドリファレンス  | ネットボランチ  DNS サービスの設定

34.7 通信タイムアウトの設定
[書式 ]
netvolante-dns  timeout  interface  time
netvolante-dns  timeout  pp time
no netvolante-dns  timeout  interface  [time]
no netvolante-dns  timeout  pp [time]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名
• [初期値 ] : -
•time
• [設定値 ] : タイムアウト秒数  (1..180)
• [初期値 ] : 90
[説明 ]
ネットボランチ  DNS サーバーとの間の通信がタイムアウトするまでの時間を秒単位で設定する。
[ノート ]
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.8 ホスト名を自動生成するか否かの設定
[書式 ]
netvolante-dns  auto  hostname  interface  [server= server_num ] switch
netvolante-dns  auto  hostname  interface  [server= server_num ] ipv6 address  [ipv6_address ]
netvolante-dns  auto  hostname  interface  [server= server_num ] ipv6 prefix  [ipv6_prefix ]
netvolante-dns  auto  hostname  pp [server= server_num ] switch
netvolante-dns  auto  hostname  pp [server= server_num ] ipv6 address  [ipv6_address ]
netvolante-dns  auto  hostname  pp [server= server_num ] ipv6 prefix  [ipv6_prefix ]
no netvolante-dns  auto  hostname  interface  [switch ]
no netvolante-dns  auto  hostname  pp [switch ]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インターフェース名、 WAN インターフェース名
• [初期値 ] : -
•server_num
• [設定値 ] :
設定値 説明
1または 2 サーバ番号
省略 省略時は 1が指定されたものとみなす
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 自動生成する
off 自動生成しない
• [初期値 ] : off
•ipv6_address
• [設定値 ] : IPv6アドレスコマンドリファレンス  | ネットボランチ  DNS サービスの設定  | 541

• [初期値 ] : -
•ipv6_prefix
• [設定値 ] : IPv6プレフィックス
• [初期値 ] : -
[説明 ]
ホスト名の自動生成機能を利用するか否かを設定する。自動生成されるホスト名は、 MACアドレス上 6桁
が"00:a0:de" のときは、 『 'y'＋(MAC アドレス下  6 桁 ).auto.netvolante.jp 』という形になる。 MACアドレス上 6桁
が"00:a0:de" 以外のときは、 『 'y'＋(MAC アドレス全  12 桁 ).auto.netvolante.jp 』という形になる。
このコマンドを  'on' に設定して、 netvolante-dns go  コマンドを実行すると、ネットボランチ  DNS サーバーから上記
のホスト名が割り当てられる。割り当てられたドメイン名は、 show status netvolante-dns  コマンドで確認することが
できる。
[ノート ]
MACアドレス上 6桁が "00:a0:de" 以外の製品に対するホスト名の自動生成には  RTX5000 / RTX3500 Rev.14.00.21 以
降、 RTX1210 Rev.14.01.16 以降のファームウェア、および、 Rev.15.02 系以降のすべてのファームウェアで対応して
いる。
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
ipv6 address および ipv6 prefix は RTX5000 / RTX3500 Rev.14.00.26 以降、 RTX1210 Rev.14.01.33 以降、 RTX830
Rev.15.02.09 以降、および、 Rev.15.04 系以降のファームウェアで使用可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.9 シリアル番号を使ったホスト名登録コマンドの設定
[書式 ]
netvolante-dns  set hostname  interface  serial
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名あるいは  "pp"
• [初期値 ] : -
[説明 ]
機器のシリアル番号を使ったホスト名を利用するためのコマンドを自動設定する。
本コマンドを実行すると、 netvolante-dns hostname host コマンドが設定される。
例えば機器のシリアル番号が D000ABCDE の場合、 netvolante-dns set hostname pp serial を実行すると、 netvolante-
dns hostname host pp server=1 SER-D000ABCDE が設定される。
[ノート ]
サブドメインをユーザが指定することはできない。
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.10 ネットボランチ  DNS サーバーの設定
[書式 ]
netvolante-dns  server  ip_address
netvolante-dns  server  name
no netvolante-dns  server  [ip_address ]
no netvolante-dns  server  [name ]
[設定値及び初期値 ]
•ip_address542 | コマンドリファレンス  | ネットボランチ  DNS サービスの設定

• [設定値 ] : IP アドレス
• [初期値 ] : -
•name
• [設定値 ] : ドメイン名
• [初期値 ] : netvolante-dns.netvolante.jp
[説明 ]
ネットボランチ  DNS サーバーの  IP アドレスまたはホスト名を設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.11 ネットボランチ DNSサーバアドレス更新機能の ON/OFF の設定
[書式 ]
netvolante-dns  server  update  address  use [server= server_num ] switch
no netvolante-dns  server  update  address  use [server= server_num ]
[設定値及び初期値 ]
•server_num
• [設定値 ] :
設定値 説明
1または 2 サーバ番号
省略 省略時は 1が指定されたものとみなす
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on サーバアドレスの更新機能を有効にする
off サーバアドレスの更新機能を停止させる
• [初期値 ] : on
[説明 ]
ネットボランチ DNSサーバからの IPアドレスの変更通知を受け取り、設定を自動更新するか否かを設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.12 ネットボランチ DNSサーバアドレス更新機能のポート番号の設定
[書式 ]
netvolante-dns  server  update  address  port [server= server_num ] port
no netvolante-dns  server  update  address  port [server= server_num ]
[設定値及び初期値 ]
•server_num
• [設定値 ] :
設定値 説明
1または 2 サーバ番号
省略 省略時は 1が指定されたものとみなす
• [初期値 ] : -
•port
• [設定値 ] : ポート番号  (1..65535)
• [初期値 ] : 2002コマンドリファレンス  | ネットボランチ  DNS サービスの設定  | 543

[説明 ]
ネットボランチ DNSサーバの IPアドレス更新通知の待ち受けポート番号を設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.13 自動更新に失敗した場合のリトライ間隔と回数の設定
[書式 ]
netvolante-dns  retry  interval  interface  interval  count
netvolante-dns  retry  interval  pp interval  count
no netvolante-dns  retry  interval  interface  [interval  count ]
no netvolante-dns  retry  interval  pp [interval  count ]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名
• [初期値 ] : -
•interval
• [設定値 ] :
• auto
•秒数  (60-300)
• [初期値 ] : auto
•count
• [設定値 ] : 回数  (1-50)
• [初期値 ] : 10
[説明 ]
ネットボランチ  DNS で自動更新に失敗した場合に、再度自動更新を行う間隔と回数を設定する。
[ノート ]
interval  に auto を設定した時には、自動更新に失敗した場合には  30 秒から  90 秒の時間をおいて再度自動更新を行
う。それにも失敗した場合には、その後、 60 秒後間隔で自動更新を試みる。
自動更新に失敗してから、指定した時間までの間に手動実行をした場合は、その後の自動更新は行われない。
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.14 ネットボランチ DNS登録の定期更新間隔の設定
[書式 ]
netvolante-dns  register  timer  [server= server_num ] time
no netvolante-dns  register  timer  [server= server_num ]
[設定値及び初期値 ]
•server_num
• [設定値 ] :
設定値 説明
1または 2 サーバ番号
省略 省略時は 1が指定されたものとみなす
• [初期値 ] : -
•time
• [設定値 ] :544 | コマンドリファレンス  | ネットボランチ  DNS サービスの設定

設定値 説明
3600 ... 2147483647 秒数
off ネットボランチ DNS登録の定期更新を行わない
• [初期値 ] : off
[説明 ]
ネットボランチ DNS登録を定期的に更新する間隔を指定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
34.15 ネットボランチ DNS の自動登録に成功したとき設定を保存するファイルの設定
[書式 ]
netvolante-dns  auto  save [server= server_num ] file
no netvolante-dns  auto  save [server= server_num ]
[設定値及び初期値 ]
•server_num
• [設定値 ] :
設定値 説明
1または 2 サーバ番号
省略 省略時は 1が指定されたものとみなす
• [初期値 ] : -
•file
• [設定値 ] :
設定値 説明
off 設定の自動保存を行わない
auto デフォルト設定ファイルに自動保存を行う
番号 自動保存を行うファイル名
• [初期値 ] : auto
[説明 ]
ネットボランチ DNS の自動登録に成功したとき、およびネットボランチ DNS サーバからのアドレス通知を受け取
ったとき、設定を自動保存するかどうか、および自動保存する場合は保存先のファイル名を指定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | ネットボランチ  DNS サービスの設定  | 545

