# 第24章: DNS の設定

> 元PDFページ: 413-426

---

第 24 章
DNS の設定
本機は、 DNS(Domain Name Service) 機能として名前解決、リカーシブサーバー機能、上位  DNS サーバーの選択機
能、簡易  DNS サーバー機能  ( 静的  DNS レコードの登録  ) を持ちます。
名前解決の機能としては、 ping  やtraceroute 、rdate、ntpdate、telnet  コマンドなどの  IP アドレスパラメータの代わ
りに名前を指定したり、 SYSLOG などの表示機能において  IP アドレスを名前解決したりします。
リカーシブサーバー機能は、 DNS サーバーとクライアントの間に入って、 DNS パケットの中継を行います。本機宛
にクライアントから届いた  DNS 問い合わせパケットを dns server  等のコマンドで設定された  DNS サーバーに中継
します。 DNS サーバーからの回答は本機宛に届くので、それをクライアントに転送します。 dns cache max entry  コ
マンドで設定した件数  (初期値  = 256) のキャッシュを持ち、キャッシュにあるデータに関しては  DNS サーバーに問
い合わせることなく返事を返すため、 DNS によるトラフィックを削減する効果があります。キャッシュは、 DNS サ
ーバーからデータを得た場合にデータに記されていた時間だけ保持されます。
DNS の機能を使用するためには、 dns server  等のコマンドで、問い合わせ先  DNS サーバーを設定しておく必要があ
ります。また、この設定は  DHCP サーバー機能において、 DHCP クライアントの設定情報にも使用されます。問い
合わせ先  DNS サーバーを設定するコマンドは複数存在しますが、 これらのうち複数のコマンドで問い合わせ先  DNS
サーバーが設定されている場合、利用できる中で最も優先順位の高いコマンドの設定が使用されます。各コマンド
による設定の優先順位は、高い順に以下の通りです。
1. dns server select  コマンド
2. dns server  コマンド
3. dns server pp  コマンド
4. dns server dhcp  コマンド
なお、これらのコマンドで問い合わせ先  DNS サーバーが全く設定されていない場合でも、 DHCP サーバーから取得
した  DNS サーバーが存在すれば、そちらが自動的に使用されます。
24.1 DNS を利用するか否かの設定
[書式 ]
dns service  service
no dns service  [service ]
[設定値及び初期値 ]
•service
• [設定値 ] :
設定値 説明
recursive DNS リカーシブサーバーとして動作する
off サービスを停止させる
• [初期値 ] : recursive
[説明 ]
DNS リカーシブサーバーとして動作するかどうかを設定する。 off を設定すると、 DNS 的機能は一切動作しない。
また、ポート  53/udp も閉じられる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.2 DNS サーバーの  IP アドレスの設定
[書式 ]
dns server  ip_address  [edns= sw] [nat46= tunnel_num ] [ip_address  [edns= sw] [nat46= tunnel_num ]...]
no dns server  [ip_address  [edns= sw] [nat46= tunnel_num ]...]
[設定値及び初期値 ]
•ip_address
• [設定値 ] : DNS サーバーの  IP アドレス  ( 空白で区切って最大  4 ヶ所まで設定可能  )コマンドリファレンス  | DNS の設定  | 413

• [初期値 ] : -
•sw
• [設定値 ] :
設定値 説明
on 対象の  DNS サーバーへの通信を  EDNS で行う
off 対象の  DNS サーバーへの通信を  DNS で行う
• [初期値 ] : off
•tunnel_num
• [設定値 ] : トンネルインターフェース番号
• [初期値 ] : -
[説明 ]
DNS サーバーの  IP アドレスを指定する。
この  IP アドレスはルーターが  DHCP サーバーとして機能する場合に  DHCP クライアントに通知するためや、 IPCP
の MS 拡張オプションで相手に通知するためにも使用される。
他のコマンドでも  DNS サーバーが設定されている場合は、最も優先順位の高いコマンドの設定が使用される。 DNS
サーバー  を設定する各種コマンドの優先順位は、本章冒頭の説明を参照。
edns オプションを省略、または  edns=off を指定すると、対象の  DNS サーバーへの名前解決は  DNS で通信を行う。
edns=on を指定すると、対象の  DNS サーバーへの名前解決は  EDNS で通信を行う。
edns=on で名前解決ができない場合、 edns=off に変更すると名前解決できる場合がある。
EDNSはバージョン  0 に対応。
nat46 オプションを指定すると  DNS46 機能が有効になり、この  DNS サーバー宛ての  A レコードの問い合わせを
AAAA レコードの問い合わせに変換する。
また、 DNS サーバーからの応答に含まれる  AAAA レコードを、 nat46 ip address pool  コマンドの設定値を使用して
A レコードに変換する。変換した  A レコードは、 DNS キャッシュに登録する。
tunnel_num  には、 tunnel translation nat46  コマンドを設定したトンネルインターフェースの番号を指定する。
[ノート ]
edns オプションは、 vRX Amazon EC2 版、 RTX5000 / RTX3500 Rev.14.00.28 以前、 RTX1210 Rev.14.01.34 以前、およ
び RTX830 Rev.15.02.13 以前では使用不可能。
nat46 オプションは、 vRX シリーズ、 RTX5000 、RTX3500 、RTX1220 Rev.15.04.03 以前、 RTX1210 、および  RTX830
Rev.15.02.19 以前では使用不可能。
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
24.3 DNS ドメイン名の設定
[書式 ]
dns domain  domain_name
no dns domain  [domain_name ]414 | コマンドリファレンス  | DNS の設定

[設定値及び初期値 ]
•domain_name
• [設定値 ] : DNS ドメインを表す文字列
• [初期値 ] : -
[説明 ]
ルーターが所属する  DNS ドメインを設定する。
ルーターのホストとしての機能  (ping,traceroute) を使うときに名前解決に失敗した場合、このドメイン名を補完して
再度解決を試みる。ルーターが  DHCP サーバーとして機能する場合、設定したドメイン名は  DHCP クライアントに
通知するためにも使用される。ルーターのあるネットワークおよびそれが含むサブネットワークの  DHCP クライア
ントに対して通知する。
空文字列を設定する場合には、 dns domain  . と入力する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.4 DNS サーバーを通知してもらう相手先情報番号の設定
[書式 ]
dns server  pp peer_num  [edns= sw]
no dns server  pp [peer_num  [edns= sw]]
[設定値及び初期値 ]
•peer_num
• [設定値 ] : DNS サーバーを通知してもらう相手先情報番号
• [初期値 ] : -
•sw
• [設定値 ] :
設定値 説明
on 対象の  DNS サーバーへの通信を  EDNS で行う
off 対象の  DNS サーバーへの通信を  DNS で行う
• [初期値 ] : off
[説明 ]
DNS サーバーを通知してもらう相手先情報番号を設定する。このコマンドで相手先情報番号が設定されていると、
DNS での名前解決を行う場合に、まずこの相手先に発信して、そこで  PPP の IPCPMS 拡張機能で通知された  DNS
サーバーに対して問い合わせを行う。
相手先に接続できなかったり、接続できても  DNS サーバーの通知がなかった場合には名前解決は行われない。
他のコマンドでも  DNS サーバーが設定されている場合は、最も優先順位の高いコマンドの設定が使用される。 DNS
サーバー  を設定する各種コマンドの優先順位は、本章冒頭の説明を参照。
edns オプションを省略、または  edns=off を指定すると、対象の  DNS サーバーへの名前解決は  DNS で通信を行う。
edns=on を指定すると、対象の  DNS サーバーへの名前解決は  EDNS で通信を行う。
edns=on で名前解決ができない場合、 edns=off に変更すると名前解決できる場合がある。
EDNSはバージョン  0 に対応。
[ノート ]
この機能を使用する場合には、 dns server pp  コマンドで指定された相手先情報に、 ppp ipcp msext  on の設定が必要
である。
edns オプションは、 vRX Amazon EC2 版、 RTX5000 / RTX3500 Rev.14.00.28 以前、 RTX1210 Rev.14.01.34 以前、およ
び RTX830 Rev.15.02.13 以前では使用不可能。
[設定例 ]
# pp select 2
pp2# ppp ipcp msext on
pp2# dns server pp 2コマンドリファレンス  | DNS の設定  | 415

[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.5 DNS サーバーアドレスを取得するインタフェースの設定
[書式 ]
dns server  dhcp  interface  [edns= sw] [nat46= tunnel_num ]
no dns server  dhcp
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• WAN インタフェース名
• vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能
•ブリッジインタフェース名
• [初期値 ] : -
•sw
• [設定値 ] :
設定値 説明
on 対象の  DNS サーバーへの通信を  EDNS で行う
off 対象の  DNS サーバーへの通信を  DNS で行う
• [初期値 ] : off
•tunnel_num
• [設定値 ] : トンネルインターフェース番号
• [初期値 ] : -
[説明 ]
DNS サーバーアドレスを取得するインタフェースを設定する。このコマンドでインタフェース名が設定されている
と、 DNS で名前解決を行うときに、指定したインタフェースで  DHCP サーバーから取得した  DNS サーバーアドレ
スに対して問い合わせを行う。 DHCP サーバーから  DNS サーバーアドレスを取得できなかった場合は名前解決を
行わない。
他のコマンドでも  DNS サーバーが設定されている場合は、最も優先順位の高いコマンドの設定が使用される。 DNS
サーバー  を設定する各種コマンドの優先順位は、本章冒頭の説明を参照。
edns オプションを省略、または  edns=off を指定すると、対象の  DNS サーバーへの名前解決は  DNS で通信を行う。
edns=on を指定すると、対象の  DNS サーバーへの名前解決は  EDNS で通信を行う。
edns=on で名前解決ができない場合、 edns=off に変更すると名前解決できる場合がある。
EDNSはバージョン  0 に対応。
nat46 オプションを指定すると  DNS46 機能が有効になり、この  DNS サーバー宛ての  A レコードの問い合わせを
AAAA レコードの問い合わせに変換する。
また、 DNS サーバーからの応答に含まれる  AAAA レコードを、 nat46 ip address pool  コマンドの設定値を使用して
A レコードに変換する。変換した  A レコードは、 DNS キャッシュに登録する。
tunnel_num  には、 tunnel translation nat46  コマンドを設定したトンネルインターフェースの番号を指定する。416 | コマンドリファレンス  | DNS の設定

[ノート ]
この機能は指定したインタフェースが  DHCP クライアントとして動作していなければならない。
ブリッジインタフェースは、 vRX Amazon EC2 版、および  RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
edns オプションは、 vRX Amazon EC2 版、 RTX5000 / RTX3500 Rev.14.00.28 以前、 RTX1210 Rev.14.01.34 以前、およ
び RTX830 Rev.15.02.13 以前では使用不可能。
nat46 オプションは、 vRX シリーズ、 RTX5000 、RTX3500 、RTX1220 Rev.15.04.03 以前、 RTX1210 、および  RTX830
Rev.15.02.19 以前では使用不可能。
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
24.6 DHCP/DHCPv6/IPCP MS 拡張で  DNS サーバーを通知する順序の設定
[書式 ]
dns notice  order  protocol  server  [server ]
no dns notice  order  protocol  [server  [server ]]
[設定値及び初期値 ]
•protocol
• [設定値 ] :
設定値 説明
dhcp DHCP による通知
dhcpv6 DHCPv6 による通知
msext IPCP MS 拡張による通知
• [初期値 ] : dhcp、dhcpv6 および  msext
•server
• [設定値 ] :
設定値 説明
none 一切通知しない
me 本機自身
server dns server  コマンドに設定したサーバー群、 protocol
に dhcpv6 を指定した場合に  DHCPv6 で割り当てら
れたサーバー群
• [初期値 ] :
• me server ( protocol  が dhcp または  msext の場合 )
• me ( protocol  が dhcpv6 の場合 )
[説明 ]
DHCP や DHCPv6 、IPCP MS 拡張では  DNS サーバーを複数通知できるが、それをどのような順序で通知するかを設
定する。コマンドリファレンス  | DNS の設定  | 417

server  に none を設定した場合、他の設定に関わらず  DNS サーバーの通知を行わなくなる。
server  に me を設定した場合、本機自身の  DNS リカーシブサーバー機能を使うことを通知する。
server  に server を設定した場合、 protocol  に dhcp または  msext を指定したときは  dns server  コマンドに設定したサー
バー群を通知し、 protocol  に dhcpv6 を指定したときは、 IPv6 網から  DHCPv6 で通知された  DNS サーバー群を通知す
る。
protocol  に dhcpv6 を指定したときは、 IPv6網と連動できない環境では、 server  の設定値にかかわらず、ルーターが
IPv6 DNS サーバーアドレスを一切通知しない。
IPCP MS 拡張では通知できるサーバーの数が最大  2 に限定されているので、 後ろに  me が続く場合は先頭の  1 つだけ
と本機自身を、 server  単独で設定されている場合には先頭の  2 つだけを通知する。
[ノート ]
dhcpv6 パラメーターは、 vRX Amazon EC2 版、 RTX5000 、RTX3510 、RTX1210 Rev.14.01.34 以前、および  RTX830
Rev.15.02.13 以前では指定不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.7 プライベートアドレスに対する問い合わせを処理するか否かの設定
[書式 ]
dns private  address  spoof  spoof
no dns private  address  spoof  [spoof ]
[設定値及び初期値 ]
•spoof
• [設定値 ] :
設定値 説明
on 処理する
off 処理しない
• [初期値 ] : off
[説明 ]
on の場合、 DNS リカーシブサーバー機能で、プライベートアドレスの  PTR レコードに対する問い合わせに対し、上
位サーバーに問い合わせを転送することなく、自分でその問い合わせに対し “NXDomain” 、すなわち「そのようなレ
コードはない」というエラーを返す。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.8 DNS サーバーへの AAAAレコードの問い合わせを制限するか否かの設定
[書式 ]
dns service  aaaa  filter  switch
no dns service  aaaa  filter  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on AAAAレコードの問い合わせを制限する
off AAAAレコードの問い合わせを制限しない
• [初期値 ] : off
[説明 ]
DNSサーバーへの AAAAレコードの問い合わせを制限するか否かを設定する。
IPv6 での接続環境がないのに  AAAAレコードが引けてしまうことで、接続に失敗するような場合は、このコマンド
により  AAAA レコードの問い合わせに対して、 AAAAレコードを回答しないようにする。
本機が DNSリレーサーバーになっている通信及び本機発の通信が影響を受ける。418 | コマンドリファレンス  | DNS の設定

[ノート ]
RTX830 は Rev.15.02.03 以降で使用可能。
RTX1210 は Rev.14.01.26 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.26 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.9 SYSLOG 表示で  DNS により名前解決するか否かの設定
[書式 ]
dns syslog  resolv  resolv
no dns syslog  resolv  [resolv ]
[設定値及び初期値 ]
•resolv
• [設定値 ] :
設定値 説明
on 解決する
off 解決しない
• [初期値 ] : off
[説明 ]
SYSLOG 表示で  DNS により名前解決するか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.10 DNS 問い合わせの内容に応じた  DNS サーバーの選択
[書式 ]
dns server  select  id server  [edns= sw] [nat46= tunnel_num ] [server2  [edns= sw] [nat46= tunnel_num ]] [type] query  [original-
sender ] [restrict pp connection-pp ]
dns server  select  id pp peer_num  [edns= sw] [default-server  [edns= sw]] [type] query  [original-sender ] [restrict pp
connection-pp ]
dns server  select  id dhcp interface  [edns= sw] [nat46= tunnel_num ] [default-server  [edns= sw] [nat46= tunnel_num ]] [type]
query  [original-sender ] [restrict pp connection-pp ]
dns server  select  id reject [ type] query  [original-sender ]
no dns server  select  id
[設定値及び初期値 ]
•id
• [設定値 ] : DNS サーバー選択テーブルの番号
• [初期値 ] : -
•server
• [設定値 ] : プライマリ  DNS サーバーの  IP アドレス
• [初期値 ] : -
•server2
• [設定値 ] : セカンダリ  DNS サーバーの  IP アドレス
• [初期値 ] : -
•type : DNS レコードタイプ
• [設定値 ] :
設定値 説明
a ホストの  IP アドレスコマンドリファレンス  | DNS の設定  | 419

設定値 説明
aaaa ホストの  IPv6 アドレス
ptr IP アドレスの逆引き用のポインタ
mx メールサーバー
ns ネームサーバー
cname 別名
any すべてのタイプにマッチする
省略 省略時は  a
• [初期値 ] : -
•query  : DNS 問い合わせの内容
• [設定値 ] :
設定値 説明
type が a、aaaa、mx、ns、cname の場合 query  はドメイン名を表す文字列であり、後方一致と
する。例えば、 "yamaha.co.jp" であれば、
rtpro.yamaha.co.jp などにマッチする。 "." を指定する
とすべてのドメイン名にマッチする。
type が ptr の場合 query  は IP アドレス  (ip_address [/masklen ]) であり、
masklen  を省略したときは  IP アドレスにのみマッチ
し、 masklen  を指定したときはネットワークアドレス
に含まれるすべての  IP アドレスにマッチする。 DNS
問い合わせに含まれる .in-addr.arpa ドメインで記述さ
れた  FQDN は、 IP アドレスへ変換された後に比較さ
れる。すべての  IP アドレスにマッチする設定はでき
ない。
reject キーワードを指定した場合 query  は完全一致とし、前方一致、及び後方一致には
"*" を用いる。つまり、前方一致では、 "NetV olante.*"
であれば、 NetV olante.jp 、NetV olante.rtpro.yamaha.co.jp
などにマッチする。また、後方一致で
は、 "*yamaha.co.jp" と記述する。
• [初期値 ] : -
•original-sender
• [設定値 ] : DNS 問い合わせの送信元の  IP アドレスの範囲
• [初期値 ] : -
•connection-pp
• [設定値 ] : DNS サーバーを選択する場合、接続状態を確認する接続相手先情報番号
• [初期値 ] : -
•peer_num
• [設定値 ] : IPCP により接続相手から通知される  DNS サーバーを使う場合の接続相手先情報番号
• [初期値 ] : -
•interface
• [設定値 ] : DHCP サーバーより取得する  DNS サーバーを使う場合の  LAN インターフェース名または  WAN イ
ンターフェース名またはブリッジインターフェース名
• [初期値 ] : -
•default-server
• [設定値 ] :
•peer_num  パラメータで指定した接続相手から  DNS サーバーを獲得できなかったときに使う  DNS サーバ
ーの  IP アドレス
•interface  パラメータで指定したインターフェースから  DNS サーバーを獲得できなかったときに使う  DNS
サーバーの  IP アドレス
• [初期値 ] : -
•sw420 | コマンドリファレンス  | DNS の設定

• [設定値 ] :
設定値 説明
on 対象の  DNS サーバーへの通信を  EDNS で行う
off 対象の  DNS サーバーへの通信を  DNS で行う
• [初期値 ] : off
•tunnel_num
• [設定値 ] : トンネルインターフェース番号
• [初期値 ] : -
[説明 ]
DNS 問い合わせの解決を依頼する  DNS サーバーとして、 DNS 問い合わせの内容および  DNS 問い合わせの送信元お
よび回線の接続状態を確認する接続相手先情報番号と  DNS サーバーとの組合せを複数登録しておき、 DNS 問い合わ
せに応じてその組合せから適切な  DNS サーバーを選択できるようにする。テーブルは小さい番号から検索され、
DNS 問い合わせの内容に query  がマッチしたら､その  DNS サーバーを用いて  DNS 問い合わせを解決しようとする。
一度マッチしたら、それ以降のテーブルは検索しない。すべてのテーブルを検索してマッチするものがない場合に
は、他のコマンドで指定された  DNS サーバーを用いる。 DNS サーバー  を設定する各種コマンドの優先順位は、本
章冒頭の説明を参照。
reject キーワードを使用した書式の場合、 query  がマッチしたら、その  DNS 問い合わせパケットを破棄し、 DNS 問い
合わせを解決しない。
restrict pp 節が指定されていると、 connection-pp  で指定した相手先がアップしているかどうかがサーバーの選択条件
に追加される。相手先がアップしていないとサーバーは選択されない。相手先がアップしていて、かつ、他の条件
もマッチしている場合に指定したサーバーが選択される。
edns オプションを省略、または  edns=off を指定すると、対象の  DNS サーバーへの名前解決は  DNS で通信を行う。
edns=on を指定すると、対象の  DNS サーバーへの名前解決は  EDNS で通信を行う。
edns=on で名前解決ができない場合、 edns=off に変更すると名前解決できる場合がある。
EDNSはバージョン  0 に対応。
nat46 オプションを指定すると  DNS46 機能が有効になり、この  DNS サーバー宛ての  A レコードの問い合わせを
AAAA レコードの問い合わせに変換する。
また、 DNS サーバーからの応答に含まれる  AAAA レコードを、 nat46 ip address pool  コマンドの設定値を使用して
A レコードに変換する。変換した  A レコードは、 DNS キャッシュに登録する。
tunnel_num  には、 tunnel translation nat46  コマンドを設定したトンネルインターフェースの番号を指定する。
[ノート ]
ブリッジインタフェースは、 vRX Amazon EC2 版、および  RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
WAN インタフェースは、 vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能。
edns オプションは、 vRX Amazon EC2 版、 RTX5000 / RTX3500 Rev.14.00.28 以前、 RTX1210 Rev.14.01.34 以前、およ
び RTX830 Rev.15.02.13 以前では指定不可能。
nat46 オプションは、 vRX シリーズ、 RTX5000 、RTX3500 、RTX1220 、RTX1210 、および  RTX830 では指定不可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•tunnel_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
•peer_numコマンドリファレンス  | DNS の設定  | 421

ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.11 静的  DNS レコードの登録
[書式 ]
ip host fqdn value  [ttl= ttl]
dns static  type name  value  [ttl= ttl]
no ip host fqdn [value ]
no dns static  type name  [value ]
[設定値及び初期値 ]
•type : 名前のタイプ
• [設定値 ] :
設定値 説明
a ホストの  IPv4 アドレス
aaaa ホストの  IPv6 アドレス
ptr IP アドレスの逆引き用のポインタ
mx メールサーバー
ns ネームサーバー
cname 別名
• [初期値 ] : -
•name、value
• [設定値 ] :
type パラメータによって以下のように意味が異なる
type パラメータ name value
a FQDN IPv4 アドレス
aaaa FQDN IPv6 アドレス
ptr IPv4 アドレス FQDN
mx FQDN FQDN
ns FQDN FQDN
cname FQDN FQDN
• [初期値 ] : -
•fqdn
• [設定値 ] : ドメイン名を含んだホスト名
• [初期値 ] : -
•ttl
• [設定値 ] : 秒数  (1～4294967295)
• [初期値 ] : -422 | コマンドリファレンス  | DNS の設定

[説明 ]
静的な  DNS レコードを定義する。
ip host  コマンドは、 dns static  コマンドで  a と ptr を両方設定することを簡略化したものである。
[ノート ]
問い合わせに対して返される  DNS レコードは以下のような特徴を持つ。
• TTL フィールドには、 ttl パラメータの設定値がセットされる。 ttl パラメータが省略された時には  1 がセットされ
る。
• Answer セクションに回答となる  DNS レコードが  1 つセットされるだけで、 Authority/Additional セクションには
DNS レコードがセットされない
• MX レコードの  preference フィールドは  0 にセットされる
[設定例 ]
# ip host pc1.rtpro.yamaha.co.jp 133.176.200.1
# dns static ptr 133.176.200.2 pc2.yamaha.co.jp
# dns static cname mail.yamaha.co.jp mail2.yamaha.co.jp
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.12 DNS 問い合わせパケットの始点ポート番号の設定
[書式 ]
dns srcport  port[-port ]
no dns srcport  [port[-port ]]
[設定値及び初期値 ]
•port
• [設定値 ] : ポート番号  (1..65535)
• [初期値 ] : 10000-10999
[説明 ]
ルーターが送信する  DNS 問い合わせパケットの始点ポート番号を設定する。
ポート番号を一つだけしか設定しなかった場合には、指定したポート番号を始点ポートとして利用する。
ポート番号を範囲で指定した場合には、 DNS 問い合わせパケットを送信するたびに、範囲内のポート番号をランダ
ムに利用する。
[ノート ]
DNS 問い合わせパケットをフィルタで扱うとき、始点番号がランダムに変化するということを考慮しておく必要が
ある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.13 DNS サーバーへアクセスできるホストの設定
[書式 ]
dns host ip_range  [ip_range ...]
dns host any
dns host none
dns host lan
no dns host
[設定値及び初期値 ]
•ip_range  : DNS サーバーへのアクセスを許可するホストの  IP アドレスまたはニーモニック
• [設定値 ] :コマンドリファレンス  | DNS の設定  | 423

設定値 説明
1 個の  IP アドレスまたは間にハイフン  (-) をはさん
だ IP アドレス  ( 範囲指定  )、およびこれらを任意に並
べたもの指定したホストからのアクセスを許可する
lanN LAN インターフェースからのアクセスを許可する
wan1 WAN インターフェースからのアクセスを許可する
( vRX シリーズ、 RTX5000 、および  RTX3500 では指
定不可能  )
bridge1 ブリッジインターフェースからのアクセスを許可す
る
vlanN VLAN インターフェースからのアクセスを許可する
( vRX シリーズ  では指定不可能  )
lanN/M タグ  VLAN インターフェースからのアクセスを許可
する
( vRX シリーズ  では指定不可能  )
• [初期値 ] : -
•any
• [設定値 ] : すべてのホストからのアクセスを許可する
• [初期値 ] : any
•none
• [設定値 ] : すべてのホストからのアクセスを禁止する
• [初期値 ] : -
•lan
• [設定値 ] : すべての  LAN 側ネットワーク内からのアクセスを許可する
• [初期値 ] : -
[説明 ]
DNS サーバー機能へのアクセスを許可するホストを設定する。
[ノート ]
このコマンドで  LAN インタフェースを指定した場合には、ネットワークアドレスと  limited broadcast address を除く
IP アドレスからのアクセスを許可する。指定した  LAN インタフェースにプライマリアドレスもセカンダリアドレ
スも設定していなければ、アクセスを許可しない。
IP アドレスとニーモニックの混在指定および複数のニーモニックの指定は、 RTX5000 、RTX3500 、および  RTX1210
Rev.14.01.10 以前では不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.14 DNS キャッシュを使用するか否かの設定
[書式 ]
dns cache  use switch
no dns cache  use [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on DNS キャッシュを利用する
off DNS キャッシュを利用しない
• [初期値 ] : on424 | コマンドリファレンス  | DNS の設定

[説明 ]
DNS キャッシュを利用するか否かを設定する。
switch  を on に設定した場合、 DNS キャッシュを利用する。すなわち、ルーターが送信した  DNS 問い合わせパケッ
トに対する上位  DNS サーバーからの返答をルーター内部に保持し、次に同じ問い合わせが発生したときでも、サー
バーには問い合わせず、キャッシュの内容を返す。
上位  DNS サーバーから得られた返答には複数の  RR レコードが含まれているが、 DNS キャッシュの保持時間は、そ
れらの  RR レコードの  TTL のうちもっとも短い時間になる。また、まったく  RR レコードが存在しない場合には、
60 秒となる。
ルーター内部に保持する  DNS エントリの数は dns cache max entry  コマンドで設定する。
switch  を off にした場合、 DNS キャッシュは利用しない。ルーターが送信した  DNS 問い合わせパケットに対する上
位 DNS サーバーからの返答はルーター内部に保持せず、同じ問い合わせがあっても毎回  DNS サーバーに問い合わ
せを行う。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.15 DNS キャッシュの最大エントリ数の設定
[書式 ]
dns cache  max entry  num
no dns cache  max entry  [num]
[設定値及び初期値 ]
•num
• [設定値 ] : 最大エントリ数  (1...1024)
• [初期値 ] : 256
[説明 ]
DNS キャッシュの最大エントリ数を設定する。
設定した数だけ、ルーター内部に  DNS キャッシュとして上位  DNS サーバーからの返答を保持できる。設定した数
を超えた場合、返答が返ってきた順で古いものから破棄される。
上位  DNS サーバーから得られた返答には複数の  RR レコードが含まれているが、 DNS キャッシュの保持時間は、そ
れらの  RR レコードの  TTL のうちもっとも短い時間になる。また、まったく  RR レコードが存在しない場合には、
60 秒となる。返答が得られてから保持時間を経過したエントリは、 DNS キャッシュから削除される。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
24.16 DNS フォールバック動作をルーター全体で統一するか否かの設定
[書式 ]
dns service  fallback  switch
no dns service  fallback  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on DNS フォールバック動作を IPv6 優先に統一する
off DNS フォールバック動作は機能ごとにまちまちであ
る
• [初期値 ] : off
[説明 ]
DNS フォールバック動作をルーターのすべての機能で統一するか否かを設定する。
DNS でホスト名を IP アドレスに変換する場合、 IPv4/IPv6 いずれかを DNS サーバーに先に問い合わせ、アドレスがコマンドリファレンス  | DNS の設定  | 425

解決できない場合に他方のアドレスを問い合わせる動作を、 DNS フォールバックと呼ぶ。ルーター自身が問い合わ
せる場合、 IPv4 を優先するか IPv6 を優先するかは機能ごとにまちまちであった。具体的には、以下の機能では DNS
フォールバック動作では IPv6 が優先されるが、その他の機能では IPv4 が優先されている。
- HTTP リビジョンアップ機能
- HTTP アップロード機能
このコマンドを on に設定すると、ルーターのすべての機能で IPv6 が優先されるようになる。
[ノート ]
DNS リカーシブサーバーとして、 LAN 内の PC 等の問い合わせを上位の DNS サーバーに転送する際には、 PC 等の
問い合わせ内容をそのまま上位サーバーに転送するため、 DNS フォールバックの動作も PC 等の実装がそのまま反
映され、このコマンドの設定には影響を受けない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830426 | コマンドリファレンス  | DNS の設定

