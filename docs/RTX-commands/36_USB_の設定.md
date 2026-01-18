# 第36章: USB の設定

> 元PDFページ: 549-549

---

第 36 章
USB の設定
36.1 USB ホスト機能を使うか否かの設定
[書式 ]
usbhost  use switch
no usbhost  use [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on USB ホスト機能を使用する
off USB ホスト機能を使用しない
• [初期値 ] : on
[説明 ]
USB ホスト機能を使用するか否かを設定する。
このコマンドが  off に設定されているときは  USB メモリをルーターに接続しても認識されない。
また、過電流により  USB ホスト機能に障害が発生した場合、 USB メモリが接続されていない状態で本コマンドを再
設定すると復旧させることができる。
[適用モデル ]
RTX3510, RTX1300, RTX1220, RTX1210, RTX840, RTX830
36.2 USB バスで過電流保護機能が働くまでの時間の設定
[書式 ]
usbhost  overcurrent  duration  duration
no usbhost  overcurrent  duration  [duration ]
[設定値及び初期値 ]
•duration
• [設定値 ] : 時間  (5..100、1単位が 10ミリ秒 )
• [初期値 ] : 5 (50ミリ秒 )
[説明 ]
過電流保護機能が働くまでの時間を設定する。ここで設定した時間、連続して過電流が検出されたら、過電流保護
機能が働く。
[適用モデル ]
RTX3510, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | USB の設定  | 549

