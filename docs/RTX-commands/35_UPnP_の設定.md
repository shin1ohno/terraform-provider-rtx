# 第35章: UPnP の設定

> 元PDFページ: 546-548

---

第 35 章
UPnP の設定
35.1 UPnP を使用するか否かの設定
[書式 ]
upnp  use use
no upnp  use
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : off
[説明 ]
UPnP 機能を使用するか否かを設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
35.2 UPnP に使用する  IP アドレスを取得するインタフェースの設定
[書式 ]
upnp  external  address  refer  interface
upnp  external  address  refer  pp peer_num
upnp  external  address  refer  default
no upnp  external  address  refer  [interface ]
no upnp  external  address  refer  pp [peer_num ]
[設定値及び初期値 ]
•interface
• [設定値 ] :
設定値 説明
LAN インタフェース名 指定した LAN インタフェースの  IP アドレスを取得
する
WAN インタフェース名 指定した WAN インタフェースの  IP アドレスを取得
する
default デフォルトルートのインタフェース
• [初期値 ] : default
•peer_num
• [設定値 ] :
•相手先情報番号
• anonymous
• [初期値 ] : -
[説明 ]
UPnP に使用する  IP アドレスを取得するインタフェースを設定する。546 | コマンドリファレンス  | UPnP の設定

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
RTX1300, RTX1220, RTX1210, RTX840, RTX830
35.3 UPnP のポートマッピング用消去タイマのタイプの設定
[書式 ]
upnp  port mapping  timer  type type
no upnp  mapping  timer  type
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
normal ARP 情報を参照しない
arp ARP 情報を参照する
• [初期値 ] : arp
[説明 ]
UPnP のポートマッピングを消去するためのタイマのタイプを設定する。
このコマンドで変更を行うと消去タイマ値は  3600 秒にセットされる。消去タイマの秒数は upnp port mapping
timer  コマンドで変更できる。
arp を指定すると upnp port mapping timer  off の設定よりも優先する。
arp に影響されずにポートマッピングを残す場合は  normal を指定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
35.4 UPnP のポートマッピングの消去タイマの設定
[書式 ]
upnp  port mapping  timer  time
no upnp  port mapping  timer
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
600..21474836 秒数
off 消去しないコマンドリファレンス  | UPnP の設定  | 547

• [初期値 ] : 3600
[説明 ]
UPnP によって生成されたポートマッピングを消去するまでの時間を設定する。
[ノート ]
upnp port mapping timer type  コマンドで設定を行った後、このコマンドを設定する。
off に設定した場合でも upnp port mapping timer type  arp の設定にしてあるとポートマッピングは消去される。
ARP がタイムアウトした状態でもポートマッピングを消去したくない場合は upnp port mapping timertype  normal
に設定するようにする。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
35.5 UPnP の syslog を出力するか否かの設定
[書式 ]
upnp  syslog  syslog
no upnp  syslog
[設定値及び初期値 ]
•syslog
• [設定値 ] :
設定値 説明
on UPnP の syslog を出力する
off UPnP の syslog を出力しない
• [初期値 ] : off
[説明 ]
UPnP の syslog を出力するか否かを設定する。デバッグレベルで出力される。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830548 | コマンドリファレンス  | UPnP の設定

