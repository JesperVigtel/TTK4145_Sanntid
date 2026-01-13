# GIT-CHEATS.md — hurtigreferanse 🚀

Kort, nyttig oversikt over vanlige Git‑kommandoer du kan bruke i repoet.

## Oppsett
- git config --global user.name "Navn"
- git config --global user.email "epost@domene.no"
- git clone <url>
- git init

## Sjekke tilstand
- git status
- git diff
- git log --oneline -n 20
- git branch -a

## Legg til og commit
- git add <fil>        # stage en fil
- git add .            # stage alle endringer
- git commit -m "Kort beskjed"
- git commit -am "Melding"  # stage+commit for endrede filer (ikke nye)

## Arbeide med branches
- git branch <navn>        # opprett
- git checkout <navn>      # bytt
- git checkout -b <navn>   # opprett + bytt
- git merge <branch>       # merge inn i nåværende
- git rebase <branch>      # rebase (bruk med forsiktighet)

## Remote (origin)
- git remote -v
- git fetch
- git pull --rebase origin <branch>
- git push origin <branch>
- git push -u origin <branch>  # sett upstream

## Avbryte / tilbakestille
- git checkout -- <fil>         # forkast lokale endringer
- git stash                      # midlertidig lagring
- git stash pop
- git reset --soft <commit>
- git reset --hard <commit>      # advarsel: sletter lokale endringer
- git revert <commit>

## Tags & release
- git tag -a v1.0 -m "release"
- git push origin v1.0

## Nyttig ved feil
- git reflog                      # finn nylige HEAD posisjoner
- git bisect                      # bisect for å finne hvilken commit som introduserte feil

## Flere tips
- Bruk SSH‑nøkkel per‑repo hvis du vil logge inn kun i denne mappen (se README eller spør meg om oppskrift).
- Hvis du jobber med andre: pull ofte og kommuniser før du pusher til delte branches.
- Ikke legg inn hemmeligheter i repo (passord, nøkler).

---

Kort eksempel: opprett branch og push
```bash
git checkout -b pc
git add .
git commit -m "Legg til GIT-CHEATS.md"
git push -u origin pc
```

Trenger du at jeg committer og pusher denne filen for deg? Si "ja, commit" så gjør jeg det (eller jeg viser kommandoene du skal kjøre).