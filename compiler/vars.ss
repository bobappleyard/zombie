(import srfi-1)
(define-record var name scope)

(define (var-bound? v)
  (let ([v     (var-name v)]
        [scope (var-scope v)])
    (any (lambda (e)
           (case (car e)
             [(lambda) (member v (cadr e))]
             [(let)    (member v (map car (cadr e)))]
             [else     #f]))
         scope)))

(define (all-vars e)
  (delete-duplicates!
   (let next ([e e] [scope '()])
     (cond
       [(symbol? e) (list (make-var e scope))]
       [(atom? e)   '()]
       [else        (let* ([scope (cons e scope)]
                           [next  (lambda (e) (next e scope))]
                           [rec   (lambda (e) (apply append '() (map next e)))])
                      (case (car e)
                            [(set! if begin) (rec (cdr e))]
                            [(lambda)        (rec (cddr e))]
                            [(let)           (append (rec (map cadr (cadr e)))
                                                     (rec (cddr e)))]
                            [else (rec e)]))]))
   equal?))

(define (invert f) (lambda args (not (apply f args))))

(define (free-vars e)
  (filter (invert var-bound?)
          (all-vars e)))

(define (boxed-vars e)
  (filter (lambda (v)
            (and (any (lambda (e) (eqv? (car e) 'set!)) (var-scope v))
                 (not (var-bound? v))))
          (all-vars e)))

(define (closures->currying e)
  (cond
    [(atom? e) e]
    [(case (car e)
       [(begin if set!) (cons (car e) (map closures->currying (cdr e)))]
       [(lambda) (let* ([free (map var-name (free-vars e)]
                        [need (filter (lambda (v) (member v free))
                                      vs)]
                        [inner (append need (cadr e))]
                        [body (map (lambda (e) (closures->currying e inner)) (cddr e))])
                   (if (null? need)
                     `(lambda ,inner ,@body)
                     `(curry (lambda ,inner ,@body) ,@need)))]
       [(let) (let ([inner (append (filter (lambda (v) (not (member v vs))) (map car (cadr e)))
                                   vs)])
                `(let ,(map (lambda (b) `(,(car b) ,(closures->currying (cadr b) vs))) (cadr e))
                  ,@(map (lambda (e) (closures->currying e inner)) (cddr e))))]
       [else (map (lambda (e) (closures->currying e vs)) e)])]))



