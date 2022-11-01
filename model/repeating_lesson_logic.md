# Repeating Lessons Logic
## Aliases
srt = start repeating time (date)
st = start time (time)
et = end time (time)
ct = current time (date time)
i = repeat inerval (interval)

## Explanation
```
a: srt + n.i + st <= ct
b: srt + n.i + et >= ct
```

A checks that the start of the lesson is before now and, b checks that the end time is after now.
**N must be an integer** as we are looking for a discrete interval.

```
1: srt + n.i + et <= ct
2: srt + n.i + et >= ct
=> 3: ct <= srt + n.i + et

Equate 3 and, 1
srt + n.i + st <= ct       <= srt + n.i + et
      n.i + st <= ct - srt <= n.i + et
      st <= ct - srt - n.i <= et

=> st <= ct - srt - n.i and ct - srt - n.i <= et
=> n1 = (ct - srt - st) / i
   n1 : Z => n1 = floor(n1)
=> n2 = (ct - srt - et) / i # We can ignore this and, check n1 for consistency with 1 and, 2.
```

Sub n1 to 1 and, 2 and check that both statements hold. If this is true then the repeating lesson
is happenening now.

