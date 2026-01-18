# 第23章: NAT 機能

> 元PDFページ: 401-412

---

第 23 章
NAT 機能
NAT 機能は、ルーターが転送する  IP パケットの始点 /終点  IP アドレスや、 TCP/UDP のポート番号を変換することに
より、アドレス体系の異なる  IP ネットワークを接続することができる機能です。
NAT 機能を用いると、プライベートアドレス空間とグローバルアドレス空間との間でデータを転送したり、 1 つの
グローバル  IP アドレスに複数のホストを対応させたりすることができます。
ヤマハルーター  では、始点 /終点  IP アドレスの変換だけを行うことを  NAT と呼び、 TCP/UDP のポート番号の変換を
伴うものを  IP マスカレードと呼んでいます。
アドレス変換規則を表す記述を  NAT ディスクリプタと呼び、それぞれの  NAT ディスクリプタには、アドレス変換
の対象とすべきアドレス空間が定義されます。アドレス空間の記述には、 nat descriptor address inner 、nat descriptor
address outer  コマンドを用います。前者は  NAT 処理の内側  (INNER) のアドレス空間を、後者は  NAT 処理の外側
(OUTER) のアドレス空間を定義するコマンドです。原則的に、これら  2 つのコマンドを対で設定することにより、
変換前のアドレスと変換後のアドレスとの対応づけが定義されます。
NAT ディスクリプタはインタフェースに対して適用されます。インタフェースに接続された先のネットワークが
NAT 処理の外側であり、 インタフェースから本機を経由して他のインタフェースから繋がるネットワークが  NAT 処
理の内側になります。
NAT ディスクリプタは動作タイプ属性を持ちます。 IP マスカレードやアドレスの静的割当てなどの機能を利用する
場合には、該当する動作タイプを選択する必要があります。
23.1 NAT 機能の動作タイプの設定
[書式 ]
nat descriptor  backward-compatibility  type
no nat descriptor  backward-compatibility  [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
1 ポートセービング  IP マスカレード機能を無効にする
2 TCP パケットに対してポートセービング  IP マスカレ
ード機能を有効にする
3 TCP パケットおよび  UDP パケットに対してポート
セービング  IP マスカレード機能を有効にする
• [初期値 ] :
• 1 (RTX5000 、RTX3500)
• 2 (上記以外 )
[説明 ]
NAT 機能全体の動作タイプを設定する。
RTX5000 / RTX3500 Rev.14.00.32 以降、および、 Rev.14.01 系以降の機種では、ポートセービング  IP マスカレード機
能に対応しており、 IP マスカレードにおいて同一のポート番号を使用して複数の接続先とのセッションを確立でき
る。  また、 type パラメーターに  3 を設定した場合は、 TCP パケットおよび  UDP パケットの両方をポートセービング
IP マスカレード機能の対象にできる。
本コマンドで type パラメーターを  1 に設定した場合の  NAT 機能の動作は、 Rev.14 系以前の、本コマンドを搭載して
いない機種及びファームウェアにおける  NAT 機能の動作と同等となる。  type パラメーターを  2 または  3 に設定し
て動作させた場合に問題が生じる場合は、 typeパラメーターを  1 にして  NAT 機能を使用する必要がある。
[ノート ]
本コマンドによる設定の変更を反映するには、ルーターの再起動が必要となる。コマンドリファレンス  | NAT 機能  | 401

type パラメーター  3 は、Rev.15.04 系以前、 vRX Amazon EC2 版 Rev.19.00.08 以前、 vRX VMware ESXi 版 Rev.19.01.09
以前、 vRX さくらのクラウド版  Rev.19.02.10 以前、 RTX1300 Rev.23.00.16 以前、 RTX3510 Rev.23.01.03 以前、およ
び、 RTX840 Rev.23.02.02 以前では指定不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.2 インタフェースへの  NAT ディスクリプタ適用の設定
[書式 ]
ip interface  nat descriptor  nat_descriptor_list  [reverse nat_descriptor_list ]
ip pp nat descriptor  nat_descriptor_list  [reverse nat_descriptor_list ]
ip tunnel  nat descriptor  nat_descriptor_list  [reverse nat_descriptor_list ]
no ip interface  nat descriptor  [nat_descriptor_list  [reverse nat_descriptor_list ]]
no ip pp nat descriptor  [nat_descriptor_list  [reverse nat_descriptor_list ]]
no ip tunnel  nat descriptor  [nat_descriptor_list  [reverse nat_descriptor_list ]]
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• WAN インタフェース名
• vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能
• [初期値 ] : -
•nat_descriptor_list
• [設定値 ] : 空白で区切られた  NAT ディスクリプタ番号  (1..2147483647) の並び  (16 個以内  )
• [初期値 ] : -
[説明 ]
適用されたインタフェースを通過するパケットに対して、リストに定義された順番で  NAT ディスクリプタによって
定義された  NAT 変換を順番に処理する。
reverse の後ろに記述した  NAT ディスクリプタでは、通常処理される  IP アドレス、ポート番号とは逆向きの  IP アド
レス、ポート番号に対して  NAT 変換を施す。
[ノート ]
LAN インタフェースの場合、 NAT ディスクリプタの外側アドレスに対しては、同一  LAN の ARP 要求に対して応答
する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.3 NAT ディスクリプタの動作タイプの設定
[書式 ]
nat descriptor  type nat_descriptor  type [hairpin= sw]
no nat descriptor  type nat_descriptor  [type [hairpin= sw]]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•type
• [設定値 ] :
設定値 説明
none NAT 変換機能を利用しない
nat 動的  NAT 変換と静的  NAT 変換を利用
masquerade 静的  NAT 変換と  IP マスカレード変換402 | コマンドリファレンス  | NAT 機能

設定値 説明
nat-masquerade 動的  NAT 変換と静的  NAT 変換と  IP マスカレード変
換
• [初期値 ] : none
•sw
• [設定値 ] :
設定値 説明
on ヘアピン NAT機能を利用する
off ヘアピン NAT機能を利用しない
• [初期値 ] : off
[説明 ]
NAT 変換の動作タイプを指定する。
[ノート ]
nat-masquerade は、動的  NAT 変換できなかったパケットを  IP マスカレード変換で救う。例えば、外側アドレスが  16
個利用可能の場合は先勝ちで  15 個 NAT 変換され、残りは  IP マスカレード変換される。
hairpin オプションは、 vRX Amazon EC2 版、 vRX VMware ESXi 版、 RTX5000 、RTX3500 、RTX1220 Rev.15.04.03 以
前、 RTX1210 、および  RTX830 Rev.15.02.23 以前では指定不可能。また、 type に none を指定したときは、 hairpin オ
プションは指定できない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.4 NAT 処理の外側  IP アドレスの設定
[書式 ]
nat descriptor  address  outer  nat_descriptor  outer_ipaddress_list
no nat descriptor  address  outer  nat_descriptor  [outer_ipaddress_list ]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•outer_ipaddress_list  : NAT 対象の外側  IP アドレス範囲のリストまたはニーモニック
• [設定値 ] :
設定値 説明
IP アドレス 1 個の  IP アドレスまたは間に  - をはさんだ  IP アドレ
ス ( 範囲指定  )、およびこれらを任意に並べたもの
ipcp PPP の IPCP の IP-Address オプションにより接続先か
ら通知される  IP アドレス
primary ip interface  address  コマンドで設定されている  IP ア
ドレス
secondary ip interface  secondary address  コマンドで設定されて
いる  IP アドレス
map-e MAP-E で自動的に生成された  IP アドレス
( vRX Amazon EC2 版、および  vRX VMware ESXi 版
では指定不可能  )
hb46pp IPv6 マイグレーション技術の国内標準プロビジョニ
ング方式により取得した  IP アドレス
( vRX シリーズ、 RTX5000 、RTX3500 、RTX1210 、お
よびノート欄記載の過去リビジョンでは指定不可
能 )コマンドリファレンス  | NAT 機能  | 403

• [初期値 ] : ipcp
[説明 ]
動的  NAT 処理の対象である外側の  IP アドレスの範囲を指定する。 IP マスカレードでは、先頭の  1 個の外側の  IP ア
ドレスが使用される。
[ノート ]
ニーモニックをリストにすることはできない。
適用されるインタフェースにより使用できるパラメータが異なる。
適用インタフェース LAN PP トンネル
ipcp × ○ ×
primary ○ × ×
secondary ○ × ×
IP アドレス ○ ○ ○
map-e × × ○
hb46pp × × ○
outer_ipaddress_list  パラメーターの  hb46pp キーワードは、 RTX3510 Rev.23.01.01 以前、 RTX1300 Rev.23.00.11 以前、
RTX1220 Rev.15.04.05 以前、 RTX830 Rev.15.02.30 以前では指定不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.5 NAT 処理の内側  IP アドレスの設定
[書式 ]
nat descriptor  address  inner  nat_descriptor  inner_ipaddress_list
no nat descriptor  address  inner  nat_descriptor  [inner_ipaddress_list ]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•inner_ipaddress_list  : NAT 対象の内側  IP アドレス範囲のリストまたはニーモニック
• [設定値 ] :
設定値 説明
IP アドレス 1 個の  IP アドレスまたは間に  - をはさんだ  IP アドレ
ス ( 範囲指定  )、およびこれらを任意に並べたもの
auto すべて
• [初期値 ] : auto
[説明 ]
NAT/IP マスカレード処理の対象である内側の  IP アドレスの範囲を指定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.6 静的  NAT エントリの設定
[書式 ]
nat descriptor  static  nat_descriptor  id outer_ip =inner_ip  [count ]
nat descriptor  static  nat_descriptor  id outer_ip =inner_ip /netmask
no nat descriptor  static  nat_descriptor  id [outer_ip =inner_ip  [count ]]404 | コマンドリファレンス  | NAT 機能

[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•id
• [設定値 ] : 静的  NAT エントリの識別情報  (1..2147483647)
• [初期値 ] : -
•outer_ip
• [設定値 ] : 外側  IP アドレス  (1 個 )
• [初期値 ] : -
•inner_ip
• [設定値 ] : 内側  IP アドレス  (1 個 )
• [初期値 ] : -
•count
• [設定値 ] :
•連続設定する個数
•省略時は  1
• [初期値 ] : -
•netmask
• [設定値 ] :
• xxx.xxx.xxx.xxx(xxx は十進数 )
• 0x に続く十六進数
•マスクビット数  (16...32)
• [初期値 ] : -
[説明 ]
NAT 変換で固定割り付けする  IP アドレスの組み合せを指定する。個数を同時に指定すると指定されたアドレスを
始点とした範囲指定とする。
[ノート ]
外側アドレスが  NAT 処理対象として設定されているアドレスである必要は無い。
静的  NAT のみを使用する場合には、 nat descriptor address outer  コマンドと nat descriptor address inner  コマンドの
設定に注意する必要がある。初期値がそれぞれ  ipcp と auto であるので、 例えば何らかの  IP アドレスをダミーで設定
しておくことで動的動作しないようにする。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.7 IP マスカレード使用時に  rlogin,rcp と ssh を使用するか否かの設定
[書式 ]
nat descriptor  masquerade  rlogin  nat_descriptor  use
no nat descriptor  masquerade  rlogin  nat_descriptor  [use]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : offコマンドリファレンス  | NAT 機能  | 405

[説明 ]
IP マスカレード使用時に  rlogin、rcp、ssh の使用を許可するか否かを設定する。
[ノート ]
on にすると、 rlogin、rcp と ssh のトラフィックに対してはポート番号を変換しなくなる。
また  on の場合に  rsh は使用できない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.8 静的  IP マスカレードエントリの設定
[書式 ]
nat descriptor  masquerade  static  nat_descriptor  id inner_ip  protocol  [outer_port =]inner_port
no nat descriptor  masquerade  static  nat_descriptor  id [inner_ip  protocol  [outer_port =]inner_port ]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•id
• [設定値 ] : 静的  IP マスカレードエントリの識別情報  (1 以上の数値  )
• [初期値 ] : -
•inner_ip
• [設定値 ] : 内側  IP アドレス  (1 個 )
• [初期値 ] : -
•protocol
• [設定値 ] :
設定値 説明
esp ESP
tcp TCP プロトコル
udp UDP プロトコル
icmp ICMP プロトコル
プロトコル番号 IANA で割り当てられている  protocol numbers
• [初期値 ] : -
•outer_port
• [設定値 ] : 固定する外側ポート番号  ( ニーモニック  )
• [初期値 ] : -
•inner_port
• [設定値 ] : 固定する内側ポート番号  ( ニーモニック  )
• [初期値 ] : -
[説明 ]
IP マスカレードによる通信でポート番号変換を行わないようにポートを固定する。
[ノート ]
outer_port  とinner_port  を指定した場合には  IP マスカレード適用時にインタフェースの外側から内側へのパケット
はouter_port  から inner_port  に、内側から外側へのパケットは inner_port  から outer_port  へとポート番号が変換され
る。
outer_port  を指定せず、 inner_port  のみの場合はポート番号の変換はされない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830406 | コマンドリファレンス  | NAT 機能

23.9 NAT の IP アドレスマップの消去タイマの設定
[書式 ]
nat descriptor  timer  nat_descriptor  time
nat descriptor  timer  nat_descriptor  protocol= protocol  [port= port_range ] time
nat descriptor  timer  nat_descriptor  tcpfin time2
no nat descriptor  timer  nat_descriptor  [time]
no nat descriptor  timer  nat_descriptor  protocol= protocol  [port= port_range ] [time]
no nat descriptor  timer  nat_descriptor  tcpfin [ time2 ]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•time
• [設定値 ] : 消去タイマの秒数  (30..21474836)
• [初期値 ] : 900
•time2
• [設定値 ] : TCP/FIN 通過後の消去タイマの秒数  (1-21474836)
• [初期値 ] : 60
•protocol
• [設定値 ] : プロトコル
• [初期値 ] : -
•port_range
• [設定値 ] : ポート番号の範囲、プロトコルが  TCP または  UDP の場合にのみ有効
• [初期値 ] : -
[説明 ]
NAT や IP マスカレードのセッション情報を保持する期間を表す  NAT タイマを設定する。 IP マスカレードの場合に
は、プロトコルやポート番号別の  NAT タイマを設定することもできる。指定されていないプロトコルの場合は、第
一の形式で設定した  NAT タイマの値が使われる。
IP マスカレードの場合には、 TCP/FIN 通過後の  NAT タイマを設定することができる。 TCP/FIN が通過したセッショ
ンは終了するセッションなので、このタイマを短くすることで  NAT テーブルの使用量を抑えることができる。
DNSの場合、このコマンドでの設定値にかかわらず、応答パケットが通過してから約 10秒でセッション情報を削除
する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.10 外側から受信したパケットに該当する変換テーブルが存在しないときの動作の設定
[書式 ]
nat descriptor  masquerade  incoming  nat_descriptor  action  [ip_address ]
no nat descriptor  masquerade  incoming  nat_descriptor
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•action
• [設定値 ] :
設定値説明
TCP/0～1023 宛てのパケット 左記以外
through 破棄して、 RST を返す 変換せずに通す
reject 破棄して、 RST を返す 破棄して、何も返さないコマンドリファレンス  | NAT 機能  | 407

設定値説明
TCP/0～1023 宛てのパケット 左記以外
discard 破棄して、何も返さない
forward 指定されたホストに転送する
• [初期値 ] : reject
•ip_address
• [設定値 ] : 転送先の  IP アドレス
• [初期値 ] : -
[説明 ]
IP マスカレードで外側から受信したパケットに該当する変換テーブルが存在しないときの動作を設定する。
action  が forward のときには ip_address  を設定する必要がある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.11 IP マスカレードで利用するポートの範囲の設定
[書式 ]
nat descriptor  masquerade  port range  nat_descriptor  port_range1  [port_range2  [port_range3  [port_range4 ]]]
no nat descriptor  masquerade  port range  nat_descriptor  [port_range1  [port_range2  [port_range3  [port_range4 ]]]]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•port_range1 、port_range2 、port_range3 、port_range4
• [設定値 ] : 間に  - をはさんだポート番号の範囲
• [初期値 ] :
Rev.14.00 系では  IP マスカレードの最大使用ポート数によって以下のように設定されている
• 4096 ：port_range1=60000-64095
• 10000 ：port_range1=60000-64095 、port_range2=54096-59999
• 20000 ：port_range1=60000-64095 、port_range2=49152-59999 、port_range3=44096-49151
• 40000 ：port_range1=60000-64095 、port_range2=49152-59999 、port_range3=24096-49151
• 65534 ：port_range1=49152-65534 、port_range2=30000-49151 、port_range3=10000-29999 、port_range4=1024-9999
Rev.14.01 系以降では  IP マスカレードの最大使用ポート数にかかわらず以下のように設定されている
• port_range1=60000-64095 、port_range2=49152-59999 、port_range3=44096-49151 ( 初期設定ポート数は  20000 )
[説明 ]
IP マスカレードで利用するポート番号の範囲を設定する。
ポート番号は、まず最初に  port_range1  の範囲から利用される。 port_range1  のポート番号がすべて使用中になった
ら、 port_range2  の範囲のポート番号を使い始める。このように、 port_range1  から  port_rangeN  の範囲まで、小さい
番号のポート範囲から順番にポート番号が利用される。
RTX5000 / RTX3500 は NAT の最大同時セッション数が  65534 であるが、初期設定ではウェルノウンポートを除いた
64511 個のポートしか使用できないため、同時セッション数を  65534 まで拡張する場合は、  本コマンドで  65534 個
のポートを使用できるようにポート範囲を広げる、あるいは  nat descriptor backward-compatibility コマンドで  type パ
ラメーターを  2 または  3 に設定する必要がある。
RTX5000 / RTX3500 Rev.14.00.32 以降、および、 Rev.14.01 系以降では、同一のポート番号を使用して複数の接続先と
のセッションを確立できるため、本コマンドで設定したポート数を超えるセッションの確立が可能である。
RTX5000 / RTX3500 Rev.14.00.32 以降、 および、 Rev.14.01 系以降では、 最大セッション数は  nat descriptor masquerade
session limit total  コマンドで設定する。  ただし  RTX5000 / RTX3500 Rev.14.00.32 以降、および、 Rev.14.01 系以降に
おいても、  nat descriptor backward-compatibility  コマンドで  type パラメーターを  1 に変更した場合は、  最大セッシ
ョン数は本コマンドで設定したポート数と同等となるため、最大セッション数を変更する場合は本コマンドの設定
を変更する必要がある。408 | コマンドリファレンス  | NAT 機能

[ノート ]
機種ごとの最大使用ポート数と利用可能なポート範囲の個数を下表に示す。
機種 最大使用ポート数 ポート範囲の個数
RTX1210 Rev.14.01.34 以降、 RTX830
Rev.15.02.10 以降、および、 Rev.15.04
系以降65534 64
RTX1210 Rev.14.01.26 以降、 RTX830
Rev.15.02.03 以降65534 16
RTX1210 Rev.14.01.20 以前、 RTX830
Rev.15.02.0165534 4
RTX5000 、RTX3500 Rev.14.00.32 以降 65534 64
RTX5000 、RTX3500 Rev.14.00.29 以前 65534 4
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.12 FTP として認識するポート番号の設定
[書式 ]
nat descriptor  ftp port nat_descriptor  port [port...]
no nat descriptor  ftp port nat_descriptor  [port...]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•port
• [設定値 ] : ポート番号  (1..65535)
• [初期値 ] : 21
[説明 ]
TCP で、このコマンドにより設定されたポート番号を  FTP の制御チャネルの通信だとみなして処理をする。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.13 IP マスカレードで変換しないポート番号の範囲の設定
[書式 ]
nat descriptor  masquerade  unconvertible  port nat_descriptor  if-possible
nat descriptor  masquerade  unconvertible  port nat_descriptor  protocol  port
no nat descriptor  masquerade  unconvertible  port nat_descriptor  protocol  [port]
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•protocol
• [設定値 ] :
設定値 説明
tcp TCP
udp UDP
• [初期値 ] : -
•port
• [設定値 ] : ポート番号の範囲コマンドリファレンス  | NAT 機能  | 409

• [初期値 ] : -
[説明 ]
IP マスカレードで変換しないポート番号の範囲を設定する。
if-possible が指定されている時には、処理しようとするポート番号が他の通信で使われていない場合には値を変換せ
ずそのまま利用する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.14 NAT のアドレス割当をログに記録するか否かの設定
[書式 ]
nat descriptor  log switch
no nat descriptor  log
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 記録する
off 記録しない
• [初期値 ] : off
[説明 ]
NAT のアドレス割当をログに記録するか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.15 SIP メッセージに含まれる  IP アドレスを書き換えるか否かの設定
[書式 ]
nat descriptor  sip nat_descriptor  sip
no nat descriptor  sip nat_descriptor
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•sip
• [設定値 ] :
設定値 説明
on 変換する
off 変換しない
auto sip use  コマンドの設定値に従う
• [初期値 ] :
• auto ( auto が指摘できる機種・リビジョン  )
• on ( 上記以外  )
[説明 ]
静的  NAT や静的  IP マスカレードで  SIP メッセージに含まれる  IP アドレスを書き換えるか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830410 | コマンドリファレンス  | NAT 機能

23.16 IP マスカレード変換時に  DF ビットを削除するか否かの設定
[書式 ]
nat descriptor  masquerade  remove  df-bit  remove
no nat descriptor  masquerade  remove  df-bit  [remove ]
[設定値及び初期値 ]
•remove
• [設定値 ] :
設定値 説明
on IP マスカレード変換時に  DF ビットを削除する
off IP マスカレード変換時に  DF ビットを削除しない
• [初期値 ] : on
[説明 ]
IP マスカレード変換時に  DF ビットを削除するか否かを設定する。
DF ビットは経路  MTU 探索のために用いるが、そのためには長すぎるパケットに対する  ICMP エラーを正しく発信
元まで返さなくてはいけない。しかし、 IP マスカレード処理では  IP アドレスなどを書き換えてしまうため、 ICMP
エラーを正しく発信元に返せない場合がある。そうなると、パケットを永遠に届けることができなくなってしまう。
このように、経路  MTU 探索のための  ICMP エラーが正しく届かない状況を、経路  MTU ブラックホールと呼ぶ。
IP マスカレード変換時に同時に  DF ビットを削除してしまうと、 この経路  MTU ブラックホールを避けることができ
る。その代わりに、経路  MTU 探索が行われないことになるので、通信効率が下がる可能性がある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.17 IP マスカレードで変換するホスト毎のセッション数の設定
[書式 ]
nat descriptor  masquerade  session  limit  nat_descriptor  id limit
no nat descriptor  masquerade  session  limit  nat_descriptor  id
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•id
• [設定値 ] : セッション数設定の識別番号  (1)
• [初期値 ] : -
•limit
• [設定値 ] :
•制限値  (1..500000) ( RTX3510 )
•制限値  (1..250000) ( RTX1300 )
•制限値  (1..150000) ( RTX840 )
•制限値  (1..65534) ( 上記以外  )
• [初期値 ] :
• 500000 ( RTX3510 )
• 250000 ( RTX1300 )
• 150000 ( RTX840 )
• 65534 ( 上記以外  )
[説明 ]
ホスト毎に  IP マスカレードで変換するセッションの最大数を設定する。
ホストはパケットの始点  IP アドレスで識別され、 任意のホストを始点とした変換テーブルの登録数が limit に制限さ
れる。コマンドリファレンス  | NAT 機能  | 411

[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
23.18 IP マスカレードで変換する合計セッション数の設定
[書式 ]
nat descriptor  masquerade  session  limit  total  nat_descriptor  limit
no nat descriptor  masquerade  session  limit  total  nat_descriptor
[設定値及び初期値 ]
•nat_descriptor
• [設定値 ] : NAT ディスクリプタ番号  (1..2147483647)
• [初期値 ] : -
•limit
• [設定値 ] :
•制限値  (1..2147483647)
• [初期値 ] :
• 500000 ( vRX シリーズ  ( 通常モード  での動作時  ) 、RTX3510 )
• 250000 ( RTX1300 )
• 150000 ( RTX840 )
• 65534 ( 上記以外  )
[説明 ]
ひとつの  NAT ディスクリプターにおいて、 IP マスカレードで変換するセッション数の最大数を設定する。  nat
descriptor masquerade session limit  コマンドとは異なり、すべてのホストのセッション数の合計が対象となる。
[ノート ]
本コマンドの設定は、 nat descriptor backward-compatibility  コマンドで、 typeパラメータを  2 または  3 に設定した場
合のみ有効となる。
RTX5000 は Rev.14.00.32 以降で使用可能。
RTX3500 は Rev.14.00.32 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830412 | コマンドリファレンス  | NAT 機能

