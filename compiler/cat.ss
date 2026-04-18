(define (main args)
  (let next ([obj (read)])
    (if (eof-object? obj)
      (print "done")
      (begin
        (write obj)
        (next (read))))))

